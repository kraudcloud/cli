package main

import (
	"fmt"

	"github.com/kraudcloud/cli/api"
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
				fmt.Fprintf(cmd.ErrOrStderr(), "error listing feeds: %v\n", err)
				return
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
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			feed, err := API().CreateFeed(cmd.Context(), api.KraudCreateFeed{
				Name:    name,
				IconURL: iconURL,
			})
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error creating feed: %v\n", err)
				return
			}

			fmt.Printf("Feed %s created\n", feed.ID)
		},
	}

	c.Flags().StringVar(&iconURL, "icon", "https://avatars.githubusercontent.com/u/97388814", "Icon URL")
	c.MarkFlagRequired("icon")

	return c
}
