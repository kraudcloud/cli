package main

import (
	"context"
	"fmt"
	"github.com/99designs/keyring"
	"github.com/kraudcloud/cli/api"
	"os"
	"strings"
	"sync"
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

			key := "token"
			if USER_CONTEXT != "default" {
				key = fmt.Sprintf("%s:%s", USER_CONTEXT, key)
			}

			i, err := kr.Get(key)
			if err != nil {

				context := ""
				if USER_CONTEXT != "default" {
					context = fmt.Sprintf(" -c %s", USER_CONTEXT)
				}

				fmt.Fprintf(os.Stderr,
					"No token available.\n"+
						"Go to https://kraudcloud.com/profile and create a token, then set with `kra%s login <token>`\n"+
						"Or set the KR_ACCESS_TOKEN environment variable.\n", context)
				os.Exit(1)
			}

			token = string(i.Data)
		}

		apiClient = api.NewClient(token)

		_, err := apiClient.GetUserMe(context.Background())
		if err != nil {
			if strings.Contains(err.Error(), "Unauthorized") {
				fmt.Fprintf(os.Stderr,
					"Invalid or expired token.\n"+
						"Go to https://kraudcloud.com/profile and create a token, then set with `kra login <token>`\n"+
						"Or set the KR_ACCESS_TOKEN environment variable.\n")
				os.Exit(1)
			}
			panic(err)
		}

	})

	return apiClient
}
