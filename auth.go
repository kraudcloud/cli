package main

import (
	"fmt"
	"github.com/99designs/keyring"
	"github.com/spf13/cobra"
)

func loginCMD() *cobra.Command {

	c := &cobra.Command{
		Use:     "login <token>",
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

			key := "token"
			if USER_CONTEXT != "default" {
				key = fmt.Sprintf("%s:%s", USER_CONTEXT, key)
			}

			err = ring.Set(keyring.Item{
				Key:  key,
				Data: []byte(args[0]),
			})

			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return c
}

func tokenCMD() *cobra.Command {

	c := &cobra.Command{
		Use:     "token",
		Aliases: []string{},
		Short:   "print auth token",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

			ring, err := keyring.Open(keyring.Config{
				ServiceName: "kraudcloud",
			})

			if err != nil {
				log.Fatal(err)
			}

			key := "token"
			if USER_CONTEXT != "default" {
				key = fmt.Sprintf("%s:%s", USER_CONTEXT, key)
			}

			item, err := ring.Get(key)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(item.Data))
		},
	}

	return c
}
