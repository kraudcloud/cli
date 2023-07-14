package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

const serviceName = "kraudcloud"

func loginCMD() *cobra.Command {

	c := &cobra.Command{
		Use:     "login <token>",
		Aliases: []string{},
		Short:   "set access token",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := "token"
			if USER_CONTEXT != "default" {
				key = fmt.Sprintf("%s:%s", USER_CONTEXT, key)
			}

			err := keyring.Set(serviceName, key, args[0])
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
			key := "token"
			if USER_CONTEXT != "default" {
				key = fmt.Sprintf("%s:%s", USER_CONTEXT, key)
			}

			item, err := keyring.Get(serviceName, key)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(item))
		},
	}

	return c
}
