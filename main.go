package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log = logrus.New()
var configDir = ""
var endpoint = "https://k8s.m.kraudcloud.com"
var format = "table"

func main() {
	configDir = getConfigDir()

	// useful for local dev
	endpointEnv := os.Getenv("KRAUD_ENDPOINT")
	if endpointEnv != "" {
		endpoint = endpointEnv
	}

	root := cobra.Command{
		Use:   "kraud [command]",
		Short: "Kraud is a CLI tool for interacting with the kraud APIs",
	}

	root.AddCommand(apps())
	root.AddCommand(feeds())
	root.AddCommand(certs())
	root.PersistentFlags().StringVar(&format, "format", "table", "Output format (table, json)")

	root.Execute()
}
