package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func feedsCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "feeds",
		Aliases: []string{"feed"},
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
		Run: func(cmd *cobra.Command, _ []string) {

			feeds, err := API().ListFeeds(cmd.Context())
			if err != nil {
				log.Fatalln(err)
			}

			table := NewTable("ID", "Name")
			for _, feed := range feeds {
				table.AddRow(
					feed.ID,
					feed.Name,
				)
			}
			table.Print()

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

			var rsp map[string]interface{}

			err = API().Do(req, &rsp)
			if err != nil {
				log.Fatalln(err)
			}

			json.NewEncoder(os.Stdout).Encode(rsp)

		},
	}

	c.Flags().StringVar(&iconURL, "icon", "https://avatars.githubusercontent.com/u/97388814", "Icon URL")
	c.MarkFlagRequired("icon")

	return c
}
