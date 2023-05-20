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
var COMPOSE_FILENAME_DEFAULT string = "docker-compose.yml"
var USER_CONTEXT string
var OUTPUT_FORMAT string

func main() {

	_, err := os.Stat("docker-compose.yml")
	if err != nil {
		COMPOSE_FILENAME_DEFAULT = "docker-compose.yaml"
	}

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
	root.AddCommand(idpCMD())
	root.AddCommand(podsCMD())
	root.AddCommand(psCMD())
	root.AddCommand(volumesCMD())

	root.PersistentFlags().StringVarP(&COMPOSE_FILENAME, "file", "f", COMPOSE_FILENAME_DEFAULT, "docker-compose.yml file")
	root.PersistentFlags().StringVarP(&USER_CONTEXT, "context", "c", "default", "user context")

	root.PersistentFlags().StringVarP(&OUTPUT_FORMAT, "output", "o", "table", "output format (table, json)")

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
