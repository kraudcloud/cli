package main

import (
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
