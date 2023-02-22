package main

import (
	"github.com/kraudcloud/cli/api"
	"os"
	"sync"
)

var createApiOnce sync.Once
var apiClient *api.Client

func API() *api.Client {

	createApiOnce.Do(func() {
		token := os.Getenv("KR_ACCESS_TOKEN")
		if token == "" {
			log.Fatal("KR_ACCESS_TOKEN is not set. Go to https://kraudcloud.com/profile to obtain a new one")
		}

		apiClient = api.NewClient(token)

	})

	return apiClient
}
