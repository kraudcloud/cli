package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func appsCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "apps",
		Aliases: []string{"apps"},
		Short:   "Manage apps",
	}

	c.AddCommand(appsPush())
	c.AddCommand(appsLs())
	c.AddCommand(appsInspect())

	return c
}

func appsLs() *cobra.Command {

	format := "table"
	feed := ""

	c := &cobra.Command{
		Use:     "ls",
		Short:   "List apps",
		Aliases: []string{"l"},
		Run: func(_ *cobra.Command, _ []string) {

			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("/apis/kraudcloud.com/v1/feeds/%s/apps", feed),
				nil,
			)
			if err != nil {
				log.Fatalln(err)
			}

			partial := struct {
				Apps []any `json:"apps"`
			}{}

			err = API().Do(req, &partial)
			if err != nil {
				log.Fatalln(err)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			err = enc.Encode(partial)
		},
	}

	c.Flags().StringVarP(&feed, "feed", "f", "", "store to push to")
	c.Flags().StringVarP(&format, "output", "o", "table", "output format (table|json)")
	c.MarkFlagRequired("feed")

	return c
}

func appsPush() *cobra.Command {
	feed := ""

	changelog := ""
	c := &cobra.Command{
		Use:     "push <app.yaml>",
		Short:   "Push an app to the kraud",
		Aliases: []string{"p"},
		Args:    cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {

			buf := &bytes.Buffer{}

			body := multipart.NewWriter(buf)
			file, err := body.CreateFormFile("template", "template.yaml")
			if err != nil {
				log.Fatalln(err)
			}

			f, err := os.Open(args[0])
			if err != nil {
				log.Fatalln(err)
			}
			defer f.Close()

			_, err = io.Copy(file, f)
			if err != nil {
				log.Fatalln(err)
			}

			if changelog != "" {
				err = body.WriteField("changelog", changelog)
				if err != nil {
					log.Fatalln(err)
				}
			}

			err = body.Close()
			if err != nil {
				log.Fatalln(err)
			}

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("/apis/kraudcloud.com/v1/feeds/%s/app", feed),
				buf,
			)
			if err != nil {
				log.Fatalln(err)
			}

			req.Header.Set("Content-Type", body.FormDataContentType())

			err = API().Do(req, nil)
			if err != nil {
				log.Fatalln(err)
			}

		},
	}

	c.Flags().StringVarP(&feed, "feed", "f", "", "store to push to")
	c.Flags().StringVarP(&changelog, "changelog", "c", "", "changelog for the app")
	c.MarkFlagRequired("feed")

	return c
}

func appsInspect() *cobra.Command {
	feed := ""

	c := &cobra.Command{
		Use:     "inspect",
		Short:   "Inspect an app",
		Aliases: []string{"i"},
		Args:    cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {

			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("/apis/kraudcloud.com/v1/feeds/%s/apps/%s/template", feed, args[0]),
				nil,
			)
			if err != nil {
				log.Fatalln(err)
			}

			var resp = map[string]interface{}{}

			err = API().Do(req, &resp)
			if err != nil {
				log.Fatalln(err)
			}

			json.NewEncoder(os.Stdout).Encode(resp)

		},
	}

	c.Flags().StringVarP(&feed, "feed", "f", "", "store to push to")
	c.MarkFlagRequired("feed")

	return c
}

type errResp struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}
