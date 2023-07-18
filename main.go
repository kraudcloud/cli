package main

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/kraudcloud/cli/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log = logrus.New()

var USER_CONTEXT string
var OUTPUT_FORMAT string

func main() {
	root := cobra.Command{
		Use:     "kra [command]",
		Short:   "kraud api command line interface",
		Version: api.Version,
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
	root.AddCommand(podLogs())

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
