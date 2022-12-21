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

func apps() *cobra.Command {
	c := &cobra.Command{
		Use:     "app",
		Aliases: []string{"apps"},
		Short:   "Manage apps",
	}

	c.AddCommand(appsPush())
	c.AddCommand(appsLs())
	c.AddCommand(appsInspect())

	return c
}

func appsLs() *cobra.Command {
	feed := ""

	c := &cobra.Command{
		Use:     "ls",
		Short:   "List apps",
		Aliases: []string{"l"},
		Run: func(_ *cobra.Command, _ []string) {
			client := authClient()

			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("%s/apis/kraudcloud.com/v1/feeds/%s/apps", endpoint, feed),
				nil,
			)
			if err != nil {
				log.Fatalln(err)
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Fatalln(err)
			}

			defer resp.Body.Close()

			if resp.StatusCode > 201 {
				log.Warnln(resp.Status)
				var e errResp
				json.NewDecoder(resp.Body).Decode(&e)
				log.WithField("error", e.Error).WithField("message", e.Message).Fatalln("Error listing apps")
				os.Exit(resp.StatusCode)
			}

			partial := struct {
				Apps []any `json:"apps"`
			}{}

			json.NewDecoder(resp.Body).Decode(&partial)

			encoded, _ := json.Marshal(partial.Apps)

			switch format {
			case "table":
				t, err := TableFromJSON(encoded)
				if err != nil {
					log.Fatalln(err)
				}
				t.Render()
			default:
				os.Stdout.Write(encoded)
			}
		},
	}

	c.Flags().StringVarP(&feed, "feed", "f", "", "store to push to")
	c.MarkFlagRequired("feed")

	return c
}

func appsPush() *cobra.Command {
	feed := ""

	changelog := ""
	c := &cobra.Command{
		Use:     "push",
		Short:   "Push an app to the kraud server",
		Aliases: []string{"p"},
		Args:    cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			client := authClient()

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
				fmt.Sprintf("%s/apis/kraudcloud.com/v1/feeds/%s/app", endpoint, feed),
				buf,
			)
			if err != nil {
				log.Fatalln(err)
			}

			req.Header.Set("Content-Type", body.FormDataContentType())

			resp, err := client.Do(req)
			if err != nil {
				log.Fatalln(err)
			}

			defer resp.Body.Close()

			if resp.StatusCode > 201 {
				log.Warnln(resp.Status)
				var e errResp
				json.NewDecoder(resp.Body).Decode(&e)
				log.WithField("error", e.Error).WithField("message", e.Message).Fatalln("Error pushing app")
				os.Exit(resp.StatusCode)
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
			client := authClient()

			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("%s/apis/kraudcloud.com/v1/feeds/%s/apps/%s/template", endpoint, feed, args[0]),
				nil,
			)
			if err != nil {
				log.Fatalln(err)
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Fatalln(err)
			}

			defer resp.Body.Close()

			if resp.StatusCode > 201 {
				log.Warnln(resp.Status)
				var e errResp
				json.NewDecoder(resp.Body).Decode(&e)
				log.WithField("error", e.Error).WithField("message", e.Message).Fatalln("Error inspecting app")
				os.Exit(resp.StatusCode)
			}

			io.Copy(os.Stdout, resp.Body)
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
