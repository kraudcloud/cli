package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"

	"fmt"
	"github.com/k0kubun/go-ansi"
	"github.com/kraudcloud/cli/api"
	"github.com/kraudcloud/cli/compose"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/colorstring"
	"github.com/schollz/progressbar/v3"
	"os"
	"runtime"
	"sync"
	"strings"

	//dockertypes "github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
)

func imagesCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "images",
		Aliases: []string{"image"},
		Short:   "Manage images",
	}

	c.AddCommand(imagesLs())
	c.AddCommand(imagePushCMD())

	return c
}

func imagesLs() *cobra.Command {

	c := &cobra.Command{
		Use:   "ls",
		Short: "List remote images",
		Run: func(cmd *cobra.Command, _ []string) {

			ls, err := API().ListImages(cmd.Context())
			if err != nil {
				panic(err)
			}

			table := NewTable("AID", "Size", "Name")
			for _, i := range ls.Items {
				if i.Amd64 == nil {
					table.AddRow(
						i.AID,
						"?",
						i.Ref,
					)
				} else {
					table.AddRow(
						i.AID,
						humanize.Bytes(uint64(i.Amd64.Size)),
						i.Ref,
					)
				}
			}
			table.Print()

		},
	}

	return c
}

type extractedFileInfo struct {
	hash     string
	tempfile string
	size     int64
}

func imageExtractFromDocker(ctx context.Context, serviceName string, ref string) (map[string]*extractedFileInfo, error) {

	var barProxy io.Writer

	if isatty.IsTerminal(os.Stdout.Fd()) {

		bar := NewBar(100, "[cyan]"+serviceName+"[reset] Extracting "+ref+" from docker")
		defer bar.Finish()

		barProxy = bar
	} else {
		colorstring.Fprintln(ansi.NewAnsiStderr(), "[cyan]"+serviceName+"[reset] Extracting "+ref+" from docker")
	}

	docker, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer docker.Close()

	img, _, err := docker.ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return nil, err
	}

	imgtar, err := docker.ImageSave(ctx, []string{img.ID})
	if err != nil {
		return nil, err
	}
	defer imgtar.Close()

	var reader io.Reader = imgtar
	if barProxy != nil {
		reader = io.TeeReader(reader, barProxy)
	}
	tr := tar.NewReader(reader)

	var tmpfiles = make(map[string]*extractedFileInfo)

	for {
		h, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		file, err := ioutil.TempFile("", "dockersave")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		hasher := sha256.New()
		w := io.MultiWriter(file, hasher)

		size, err := io.Copy(w, tr)
		if err != nil {
			return nil, err
		}

		tmpfiles[h.Name] = &extractedFileInfo{
			hash:     fmt.Sprintf("%x", hasher.Sum(nil)),
			tempfile: file.Name(),
			size:     size,
		}
	}

	return tmpfiles, nil
}

func uploadLayers(serviceName string, r map[string]*extractedFileInfo) error {

	total := int64(0)
	for _, v := range r {
		total += v.size
	}

	var bar *progressbar.ProgressBar
	if isatty.IsTerminal(os.Stdout.Fd()) {
		bar = NewBar(int(total), "[cyan]"+serviceName+"[reset] Uploading layers ")
		defer bar.Finish()
	} else {
		colorstring.Fprintln(ansi.NewAnsiStderr(), "[cyan]"+serviceName+"[reset] Uploading layers")
	}

	var wg sync.WaitGroup

	for _, v := range r {

		wg.Add(1)

		go func(v *extractedFileInfo) {

			defer wg.Done()

			layertargz, err := os.Open(v.tempfile)
			if err != nil {
				panic(err)
			}
			defer layertargz.Close()

			var reader io.Reader = layertargz
			if bar != nil {
				reader = io.TeeReader(reader, bar)
			}

			hasher := sha256.New()

			pr, pw := io.Pipe()

			go func() {
				defer pw.Close()
				zipper := gzip.NewWriter(io.MultiWriter(pw, hasher))
				defer zipper.Close()
				io.Copy(zipper, reader)
			}()

			pushedlayer, err := API().PushLayer(context.Background(),
				"sha256:"+v.hash,
				pr,
				uint64(v.size),
			)
			if err != nil {

				if strings.Contains(err.Error(), "Conflict") {
					return
				}


				fmt.Println()
				panic(err)
			}

			if pushedlayer.Sha256 != fmt.Sprintf("%x", hasher.Sum(nil)) {
				// this also happens when we close early because the remove already has the layer
				// panic("sha256 mismatch after upload")
			}

		}(v)

	}

	wg.Wait()

	return nil
}

func imagePushCMD() *cobra.Command {

	c := &cobra.Command{
		Use:   "push",
		Short: "Push local images to the kraud",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

			spec, err := compose.ParseFile(COMPOSE_FILENAME)
			if err != nil {
				panic(err)
			}

			i := 0
			for serviceName, s := range spec.Services {
				if i > 0 {
					fmt.Println()
				}
				i++

				ref := s.Image

				colorstring.Fprintln(ansi.NewAnsiStderr(), "[cyan]"+serviceName+"[reset] Analyzing image "+ref)

				docker, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
				if err != nil {
					panic(err)
				}
				defer docker.Close()

				// first get the state of the remote image
				remoteImage, _ := API().InspectImage(cmd.Context(), ref)

				// then get the state of the local image
				localImage, _, _ := docker.ImageInspectWithRaw(cmd.Context(), ref)

				// if both exist and are valid, do nothing
				if remoteImage != nil && localImage.ID != "" {
					if localImage.ID == remoteImage.Amd64.OciID {
						colorstring.Fprintln(ansi.NewAnsiStderr(), "[cyan]"+serviceName+"[reset] Remote image is up to date")
						fmt.Println(remoteImage.AID)
						continue
					}
				}

				// if only the remote exists, do nothing
				if remoteImage != nil && localImage.ID == "" {
					colorstring.Fprintln(ansi.NewAnsiStderr(), "[cyan]"+serviceName+"[reset] Image "+ref+" not available locally!")
					fmt.Println(remoteImage.AID)
					continue
				}

				files, err := imageExtractFromDocker(cmd.Context(), serviceName, ref)

				defer func() {
					for _, t := range files {
						os.Remove(t.tempfile)
					}
				}()

				if err != nil {
					panic(err)
				}

				if files["manifest.json"] == nil {
					panic("manifest.json not found")
				}

				var manifest []struct {
					Config string
					Layers []string
				}

				manifestFile, err := os.Open(files["manifest.json"].tempfile)
				if err != nil {
					panic(err)
				}

				if err := json.NewDecoder(manifestFile).Decode(&manifest); err != nil {
					panic(err)
				}

				var config struct {
					Architecture string          `json:"architecture"`
					Config       json.RawMessage `json:"config"`
					Rootfs       struct {
						Type    string   `json:"type"`
						DiffIDs []string `json:"diff_ids"`
					} `json:"rootfs"`
				}

				configString, err := ioutil.ReadFile(files[manifest[0].Config].tempfile)
				if err != nil {
					panic(err)
				}

				if err := json.Unmarshal(configString, &config); err != nil {
					panic(err)
				}

				ociid := "sha256:" + files[manifest[0].Config].hash

				layers := make(map[string]*extractedFileInfo)
				for _, m := range manifest {
					for _, l := range m.Layers {
						layers[l] = files[l]
						if files[l] == nil {
							panic(fmt.Sprintf("layer missing %s", l))
						}
					}
				}

				err = uploadLayers(serviceName, layers)
				if err != nil {
					panic(err)
				}

				colorstring.Fprintln(ansi.NewAnsiStderr(), "[cyan]"+serviceName+"[reset] Creating references")

				layerRefs := []api.KraudLayerReference{}
				for _, diffID := range config.Rootfs.DiffIDs {
					var diffID = diffID
					layerRefs = append(layerRefs, api.KraudLayerReference{
						OciID: &diffID,
					})
				}

				rsp, err := API().CreateImage(context.Background(), api.CreateImageJSONBody{
					Ref:          ref,
					Config:       string(config.Config),
					OciID:        ociid,
					Architecture: runtime.GOARCH,
					Layers:       layerRefs,
				})

				if err != nil {
					panic(err)
				}

				for _, rn := range rsp.Renamed {
					colorstring.Fprintln(ansi.NewAnsiStderr(), "[cyan]"+serviceName+"[reset] Renamed existing image to "+rn.Ref)
				}

				fmt.Println(rsp.Created.AID)

			}

		},
	}

	return c

}
