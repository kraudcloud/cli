package main

import (
	"github.com/kraudcloud/cli/compose"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func upCMD() *cobra.Command {
	logLevel := 0
	force := false

	c := &cobra.Command{
		Use:    "up",
		Short:  "up app",
		Long:   "up takes a docker-compose file as input and ups the app on kraud cloud. Both it's API and functions are unstable, use at your own risk.",
		Hidden: true, // TODO: remove this when it's stable

		PreRun: func(_ *cobra.Command, _ []string) { logLevel += 3 },
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := compose.ParseFile(COMPOSE_FILENAME)
			if err != nil {
				return err
			}

			options := []func(*compose.UpOptions){
				compose.WithLevel(logrus.Level(logLevel)),
			}
			if force {
				options = append(options, compose.WithForce())
			}

			err = compose.Up(cmd.Context(), DockerClient(), project, options...)
			if err != nil {
				return err
			}

			log.Info("up success")
			return nil
		},
	}

	c.Flags().CountVarP(&logLevel, "v", "v", "verbose")
	c.Flags().BoolVarP(&force, "force", "f", false, "force recreate")

	return c
}
