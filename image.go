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
				log.Fatalln(err)
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

func imageExtractFromDocker(ref string) (map[string]string, error) {

	// docker save will save all matching images, so get exact id and size here

	cmd := exec.Command("docker", "inspect", ref, "--format", "{{.Id}} {{.Size}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
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
		log.Fatalln(err)
	}
	defer stdout.Close()

	if err := docker.Start(); err != nil {
		log.Fatalln(err)
	}

	var reader io.Reader = stdout
	if isatty.IsTerminal(os.Stdout.Fd()) {

		bar := NewBar(int(expectedsize), "[cyan][1/4][reset] Extracting image from docker")
		defer bar.Finish()

		reader = io.TeeReader(reader, bar)

	} else {
		log.Println("Extracting image from docker ...")
	}

	tr := tar.NewReader(reader)

	var tmpfiles = make(map[string]string)

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
			log.Fatal(err)
		}
		defer file.Close()

		if _, err := io.Copy(file, tr); err != nil {
			return nil, err
		}

		tmpfiles[h.Name] = file.Name()
	}

	return tmpfiles, nil
}

type compressedInfo struct {
	shaBeforeZip string
	shaAfterZip  string
	tmpfilename  string
	sizeAfterZip int64
}

func compressLayers(files map[string]string) (r []compressedInfo, err error) {

	var total int64

	for _, tmpname := range files {
		f, err := os.Open(tmpname)
		if err != nil {
			return r, err
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return r, err
		}

		total += fi.Size()
	}

	var bar *progressbar.ProgressBar
	if isatty.IsTerminal(os.Stdout.Fd()) {
		bar = NewBar(int(total), "[cyan][2/4][reset] Compressing layers ")
		defer bar.Finish()
	} else {
		log.Println("Compressing Layers ...")
	}

	var wg sync.WaitGroup
	var lock sync.Mutex

	for _, tmpname := range files {

		wg.Add(1)

		go func(tmpname string) {

			fi, err := os.Open(tmpname)
			if err != nil {
				log.Fatalln(err)
			}
			defer fi.Close()

			os.Remove(tmpname)

			fo, err := os.Create(tmpname)
			if err != nil {
				log.Fatalln(err)
			}
			defer fo.Close()

			h1 := sha256.New()
			h2 := sha256.New()

			zipper := gzip.NewWriter(io.MultiWriter(fo, h1))

			oo := io.MultiWriter(zipper, h2)

			if bar != nil {
				oo = io.MultiWriter(oo, bar)
			}

			if _, err := io.Copy(oo, fi); err != nil {
				log.Fatalln(err)
			}

			zipper.Close()

			foi, err := fo.Stat()
			if err != nil {
				log.Fatalln(err)
			}

			lock.Lock()
			defer lock.Unlock()

			r = append(r, compressedInfo{
				shaBeforeZip: fmt.Sprintf("%x", h2.Sum(nil)),
				shaAfterZip:  fmt.Sprintf("%x", h1.Sum(nil)),
				tmpfilename:  tmpname,
				sizeAfterZip: foi.Size(),
			})

			wg.Done()

		}(tmpname)
	}

	wg.Wait()

	return r, nil
}

func uploadLayers(r []compressedInfo) error {

	total := int64(0)
	for _, v := range r {
		total += v.sizeAfterZip
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

		go func(v compressedInfo) {

			defer wg.Done()

			layertargz, err := os.Open(v.tmpfilename)
			if err != nil {
				log.Fatalln(err)
			}
			defer layertargz.Close()

			var reader io.Reader = layertargz
			if bar != nil {
				reader = io.TeeReader(reader, bar)
			}

			_, err = API().PushLayer(context.Background(),
				"sha256:"+v.shaBeforeZip,
				reader,
				uint64(v.sizeAfterZip),
				v.shaAfterZip,
			)
			if err != nil {
				log.Fatalln(err)
			}
		}(v)

	}

	wg.Wait()

	return nil
}

func imagePushCMD() *cobra.Command {

	c := &cobra.Command{
		Use:   "push [ref]",
		Short: "Push a local docker image to the kraud",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			files, err := imageExtractFromDocker(args[0])

			defer func() {
				for _, t := range files {
					os.Remove(t)
				}
			}()

			if err != nil {
				log.Fatalln(err)
			}

			if files["manifest.json"] == "" {
				log.Fatalln("manifest.json not found")
			}

			var manifest []struct {
				Config string
				Layers []string
			}

			manifestFile, err := os.Open(files["manifest.json"])
			if err != nil {
				log.Fatalln(err)
			}

			if err := json.NewDecoder(manifestFile).Decode(&manifest); err != nil {
				log.Fatalln(err)
			}

			var config struct {
				Architecture string                 `json:"architecture"`
				Config       map[string]interface{} `json:"config"`
				Rootfs       struct {
					Type    string   `json:"type"`
					DiffIDs []string `json:"diff_ids"`
				} `json:"rootfs"`
			}

			configString, err := ioutil.ReadFile(files[manifest[0].Config])
			if err != nil {
				log.Fatalln(err)
			}

			if err := json.Unmarshal(configString, &config); err != nil {
				log.Fatalln(err)
			}

			configHash := sha256.Sum256(configString)

			layers := make(map[string]string)
			for _, m := range manifest {
				for _, l := range m.Layers {
					layers[l] = files[l]
					if files[l] == "" {
						log.Fatalln("layer missing", l)
					}
				}
			}

			ci, err := compressLayers(layers)
			if err != nil {
				log.Fatalln(err)
			}

			err = uploadLayers(ci)
			if err != nil {
				log.Fatalln(err)
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
				Ref:          args[0],
				Config:       config.Config,
				OciID:        fmt.Sprintf("sha256:%x", configHash[:]),
				Architecture: runtime.GOARCH,
				Layers:       layerRefs,
			})

			if err != nil {
				log.Fatalln(err)
			}

			for _, rn := range rsp.Renamed {
				fmt.Fprintln(os.Stderr, "Renamed existing image to", rn.Ref)
			}

			fmt.Println(rsp.Created.AID)

		},
	}

	return c

}
