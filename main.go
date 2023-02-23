package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log = logrus.New()

func main() {

	root := cobra.Command{
		Use:   "kra [command]",
		Short: "kraud api command line interface",
	}

	root.AddCommand(feedsCMD())
	root.AddCommand(appsCMD())
	root.AddCommand(loginCMD())
	root.AddCommand(domainsCMD())
	root.AddCommand(usersCMD())
	root.AddCommand(imagesCMD())
	root.AddCommand(layersCMD())
	root.AddCommand(imagePushCMD())

	root.Execute()
}
