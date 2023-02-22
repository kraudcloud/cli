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
		Use:     "feeds",
		Aliases: []string{"feeds"},
		Short:   "Manage feeds",
	}

	c.AddCommand(feedsLs())
	c.AddCommand(feedCreate())

	return c
}

func feedsLs() *cobra.Command {

	var format string

	c := &cobra.Command{
		Use:   "ls",
		Short: "List feeds",
		Run: func(_ *cobra.Command, _ []string) {

			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("/apis/kraudcloud.com/v1/feeds"),
				nil,
			)
			if err != nil {
				log.Fatalln(err)
			}

			switch format {
			case "table":
				req.Header.Set("Accept", "application/json;as=Table")
			default:
				req.Header.Set("Accept", "application/json")
			}

			resp, err := Do(req)
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
			default:
				os.Stdout.Write(encoded)
			}
		},
	}
	c.Flags().StringVarP(&format, "output", "o", "table", "output format (table|json)")

	return c
}

func feedCreate() *cobra.Command {
	iconURL := ""
	format := "table"

	c := &cobra.Command{
		Use:     "create",
		Aliases: []string{"new"},
		Short:   "Create feed",
		Args:    cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {

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
				fmt.Sprintf("/apis/kraudcloud.com/v1/feeds"),
				buf,
			)
			if err != nil {
				log.Fatalln(err)
			}

			req.Header.Add("Content-Type", "application/json")

			switch format {
			case "table":
				req.Header.Set("Accept", "application/json;as=Table")
			default:
				req.Header.Set("Accept", "application/json")
			}

			resp, err := Do(req)
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

			switch format {
			default:
				io.Copy(os.Stdout, resp.Body)
			}
		},
	}

	c.Flags().StringVar(&iconURL, "icon", "https://avatars.githubusercontent.com/u/97388814", "Icon URL")
	c.MarkFlagRequired("icon")
	c.Flags().StringVarP(&format, "output", "o", "table", "output format (table|json)")

	return c
}
