package main

import (
	"fmt"

	"github.com/99designs/keyring"
	"github.com/spf13/cobra"
)

var openKeyring = func() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName: "kraudcloud",
		// Only for macos, as a workaround when not built using cgo
		// TODO(@Karitham): Find a good way to have the action build cgo binaries for macos
		FileDir:          FileDir,
		FilePasswordFunc: keyring.TerminalPrompt,
	})
}

func loginCMD() *cobra.Command {

	c := &cobra.Command{
		Use:     "login <token>",
		Aliases: []string{},
		Short:   "set access token",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ring, err := openKeyring()
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

			ring, err := openKeyring()
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
