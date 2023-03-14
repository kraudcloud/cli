package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"runtime/debug"
	"strings"
)

var log = logrus.New()

var COMPOSE_FILENAME string
var USER_CONTEXT string

func main() {

	root := cobra.Command{
		Use:     "kra [command]",
		Short:   "kraud api command line interface",
		Version: "1.1.2",
	}

	root.AddCommand(feedsCMD())
	root.AddCommand(appsCMD())
	root.AddCommand(loginCMD())
	root.AddCommand(domainsCMD())
	root.AddCommand(usersCMD())
	root.AddCommand(imagesCMD())
	root.AddCommand(layersCMD())
	root.AddCommand(imagePushCMD())
	root.AddCommand(setupCMD())
	root.AddCommand(eventsCMD())
	root.AddCommand(tokenCMD())

	root.PersistentFlags().StringVarP(&COMPOSE_FILENAME, "file", "f", "docker-compose.yml", "docker-compose.yml file")
	root.PersistentFlags().StringVarP(&USER_CONTEXT, "context", "c", "default", "user context")

	defer func() {
		if r := recover(); r != nil {

			if err, ok := r.(error); ok {
				if strings.Contains(err.Error(), "runtime error:") {
					log.Error(r)
					debug.PrintStack()
					os.Exit(1)
				}
			}

			log.Fatal(r)

		}
	}()

	root.Execute()
}
