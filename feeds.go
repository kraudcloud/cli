package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func feeds() *cobra.Command {
	c := &cobra.Command{
		Use:     "feed",
		Aliases: []string{"feeds"},
		Short:   "Manage feeds",
	}

	c.AddCommand(feedsLs())
	c.AddCommand(feedCreate())

	return c
}

func feedsLs() *cobra.Command {
	c := &cobra.Command{
		Use:   "ls",
		Short: "List feeds",
		Run: func(_ *cobra.Command, _ []string) {
			client := authClient()

			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("%s/apis/kraudcloud.com/v1/feeds", endpoint),
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
				log.WithField("error", e.Error).WithField("message", e.Message).Fatalln("Error listing feeds")
				os.Exit(resp.StatusCode)
			}

			encoded, _ := io.ReadAll(resp.Body)

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

	return c
}

func feedCreate() *cobra.Command {
	iconURL := ""

	c := &cobra.Command{
		Use:     "create",
		Aliases: []string{"new"},
		Short:   "Create feed",
		Args:    cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			client := authClient()

			name := args[0]

			b := struct {
				Name    string `json:"name"`
				IconURL string `json:"icon_url"`
			}{
				Name:    name,
				IconURL: iconURL,
			}

			buf := &bytes.Buffer{}
			json.NewEncoder(buf).Encode(b)

			req, err := http.NewRequest(
				"POST",
				fmt.Sprintf("%s/apis/kraudcloud.com/v1/feeds", endpoint),
				buf,
			)
			if err != nil {
				log.Fatalln(err)
			}

			req.Header.Add("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				log.Fatalln(err)
			}

			defer resp.Body.Close()

			if resp.StatusCode > 201 {
				log.Warnln(resp.Status)
				var e errResp
				json.NewDecoder(resp.Body).Decode(&e)
				log.WithField("error", e.Error).WithField("message", e.Message).Fatalln("Error creating feed")
				os.Exit(resp.StatusCode)
			}

			io.Copy(os.Stdout, resp.Body)
		},
	}

	c.Flags().StringVar(&iconURL, "icon", "https://avatars.githubusercontent.com/u/97388814", "Icon URL")
	c.MarkFlagRequired("icon")

	return c
}
