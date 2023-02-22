package main

import (
	"github.com/99designs/keyring"
	"github.com/spf13/cobra"
)

func auth() *cobra.Command {

	c := &cobra.Command{
		Use:     "auth <token>",
		Aliases: []string{},
		Short:   "set access token",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			ring, err := keyring.Open(keyring.Config{
				ServiceName: "kraudcloud",
			})

			if err != nil {
				log.Fatal(err)
			}

			err = ring.Set(keyring.Item{
				Key:  "token",
				Data: []byte(args[0]),
			})

			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return c
}
