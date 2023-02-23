package main

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"os"
)

func usersCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "users",
		Aliases: []string{"user"},
		Short:   "Manage users",
	}

	c.AddCommand(usersMe())

	return c
}

func usersMe() *cobra.Command {

	c := &cobra.Command{
		Use:   "me",
		Short: "information about the caller",
		Run: func(cmd *cobra.Command, _ []string) {

			me, err := API().GetUserMe(cmd.Context())
			if err != nil {
				log.Fatalln(err)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(me)

		},
	}

	return c
}
