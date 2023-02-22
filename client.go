package main

import (
	"github.com/99designs/keyring"
	"github.com/kraudcloud/cli/api"
	"sync"
	"os"
	"fmt"
)

var createApiOnce sync.Once
var apiClient *api.Client

func API() *api.Client {

	createApiOnce.Do(func() {

		token := os.Getenv("KR_ACCESS_TOKEN")
		if token == "" {

			kr, err := keyring.Open(keyring.Config{
				ServiceName: "kraudcloud",
			})
			if err != nil {
				log.Fatal(err)
			}

			i, err := kr.Get("token")
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"No token available.\n"+
					"Go to https://kraudcloud.com/profile and create a token, then set with `kra auth <token>`\n" +
					"Or set the KR_ACCESS_TOKEN environment variable.\n")
				os.Exit(1)
			}

			token = string(i.Data)
		}

		//TODO verify token is still valid

		apiClient = api.NewClient(token)

	})

	return apiClient
}
