package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/kraudcloud/cli/api"
	"github.com/zalando/go-keyring"
)

var createApiOnce sync.Once
var apiClient *api.Client

func API() *api.Client {
	createApiOnce.Do(initAPI)
	return apiClient
}

func initAPI() {
	token, err := getToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	baseURL, err := getBaseURL()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	apiClient = api.NewClient(token, baseURL)

	if err = getMe(context.Background(), apiClient, token); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func getToken() (string, error) {
	token := os.Getenv("KR_ACCESS_TOKEN")
	if token != "" {
		return token, nil
	}

	key := "token"
	if USER_CONTEXT != "default" {
		key = fmt.Sprintf("%s:%s", USER_CONTEXT, key)
	}

	token, err := keyring.Get(serviceName, key)
	if err == nil {
		return token, nil
	}

	context := ""
	if USER_CONTEXT != "default" {
		context = fmt.Sprintf(" -c %s", USER_CONTEXT)
	}

	return "", fmt.Errorf(`no token available.
	Go to https://kraudcloud.com/profile and create a token, then set with `+"`kra%s login <token>`"+`
	Or set the KR_ACCESS_TOKEN environment variable`, context)
}

func getMe(ctx context.Context, apiClient *api.Client, token string) error {
	_, err := apiClient.GetUserMe(ctx)
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "Unauthorized") {
		return fmt.Errorf(`ivalid or expired token.
Go to https://kraudcloud.com/profile and create a token, then set with `+"`kra login <token>`"+`
Or set the KR_ACCESS_TOKEN environment variable.\n\n%w`, err)
	}
	panic(err)
}

func getBaseURL() (*url.URL, error) {
	baseURL := &url.URL{
		Scheme: "https",
		Host:   "api.kraudcloud.com",
	}

	host := os.Getenv("KR_HOST")
	if host == "" {
		return baseURL, nil
	}

	return url.Parse(host)
}
