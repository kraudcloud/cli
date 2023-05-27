package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	dclient "github.com/docker/docker/client"
	"github.com/kraudcloud/cli/api"
)

var createApiOnce sync.Once
var apiClient *api.Client

func API() *api.Client {

	createApiOnce.Do(func() {

		token := os.Getenv("KR_ACCESS_TOKEN")
		if token == "" {

			kr, err := openKeyring()
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

func DockerClient() *dclient.Client {
	ctx, err := dockerContextInspect(USER_CONTEXT)
	if err != nil {
		log.Fatalln(err)
	}

	docker, err := dclient.NewClientWithOpts(
		dclient.FromEnv,
		dclient.WithAPIVersionNegotiation(),
		dclient.WithHost(ctx.Endpoints.Docker.Host),
		dclient.WithTLSClientConfig(
			path.Join(ctx.Storage.TLSPath, "docker", ctx.TLSMaterial.Docker[0]),
			path.Join(ctx.Storage.TLSPath, "docker", ctx.TLSMaterial.Docker[1]),
			path.Join(ctx.Storage.TLSPath, "docker", ctx.TLSMaterial.Docker[2]),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	return docker
}
