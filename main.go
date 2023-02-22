package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log = logrus.New()

var outputFormat string

func main() {

	root := cobra.Command{
		Use:   "kra [command]",
		Short: "kraud api command line interface",
	}

	root.AddCommand(apps())
	root.AddCommand(feeds())
	root.AddCommand(certs())

	root.Execute()
}
