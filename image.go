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
	"os"
	"os/exec"

	"fmt"
	"github.com/k0kubun/go-ansi"
	"github.com/kraudcloud/cli/api"
	"github.com/kraudcloud/cli/compose"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/colorstring"
	"github.com/schollz/progressbar/v3"
	"runtime"
	"strconv"
	"strings"
	"sync"
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

func imageExtractFromDocker(ref string) (map[string]*extractedFileInfo, error) {

	// docker save will save all matching images, so get exact id and size here

	cmd := exec.Command("docker", "image", "inspect", "--format", "{{.Id}} {{.Size}}", ref)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker image inspect --format '{{.Id}} {{.Size}}' '%s': %w", ref, err)
	}
	outs := strings.Split(string(strings.TrimSpace(string(out))), " ")
	expectedsize, err := strconv.Atoi(outs[1])
	if err != nil {
		return nil, err
	}

	docker := exec.Command("docker", "save", outs[0])
	docker.Stderr = os.Stderr

	stdout, err := docker.StdoutPipe()
	if err != nil {
		panic(err)
	}
	defer stdout.Close()

	if err := docker.Start(); err != nil {
		panic(err)
	}

	var reader io.Reader = stdout
	if isatty.IsTerminal(os.Stdout.Fd()) {

		bar := NewBar(int(expectedsize), "[cyan][1/4][reset] Extracting "+ref+" from docker")
		defer bar.Finish()

		reader = io.TeeReader(reader, bar)

	} else {
		log.Println("Extracting " + ref + " from docker ...")
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

func uploadLayers(r map[string]*extractedFileInfo) error {

	total := int64(0)
	for _, v := range r {
		total += v.size
	}

	var bar *progressbar.ProgressBar
	if isatty.IsTerminal(os.Stdout.Fd()) {
		bar = NewBar(int(total), "[cyan][3/4][reset] Uploading layers ")
		defer bar.Finish()
	} else {
		log.Println("Uploading Layers ...")
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

			for _, s := range spec.Services {

				ref := s.Image

				files, err := imageExtractFromDocker(ref)

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

				err = uploadLayers(layers)
				if err != nil {
					panic(err)
				}

				colorstring.Fprintln(ansi.NewAnsiStderr(), "[cyan][4/4][reset] Creating references")

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
					fmt.Fprintln(os.Stderr, "Renamed existing image to", rn.Ref)
				}

				fmt.Println(rsp.Created.AID)

			}

		},
	}

	return c

}
