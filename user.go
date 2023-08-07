package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
				fmt.Fprintf(cmd.ErrOrStderr(), "error getting user: %v\n", err)
				return
			}

			identJSONEncoder(os.Stdout, me)
		},
	}

	return c
}
