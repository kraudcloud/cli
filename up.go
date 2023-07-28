package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/kraudcloud/cli/api"
	"github.com/kraudcloud/cli/compose"
	"github.com/spf13/cobra"
)

func UpCMD() *cobra.Command {
	namespace := "default"
	env := map[string]string{}
	envFile := ".env"

	c := &cobra.Command{
		Use:   "up [docker-compose file]",
		Short: "run an application",
		RunE: func(cmd *cobra.Command, args []string) error {
			file := dockerComposeFile()
			if len(args) > 0 {
				file = args[0]
			}

			template, err := os.ReadFile(file)
			if err != nil {
				return err
			}

			// needed env neededVars
			neededVars := compose.GetTemplateVars(bytes.NewReader(template))
			loaders := []compose.EnvLoader{
				compose.LoadKV(env),
				compose.LoadKVSlice(os.Environ()),
			}

			// load env vars from file
			if envFile != "" {
				f, err := os.Open(envFile)
				switch {
				case os.IsNotExist(err):
					break
				case err != nil:
					return err
				}

				defer f.Close()

				loaders = append(loaders, compose.LoadEnvReader(f))
			}

			// load env vars
			env, err := compose.LoadEnv(neededVars, loaders...)
			if err != nil {
				return err
			}

			detach, _ := cmd.Flags().GetBool("detach")

			err = API().Launch(cmd.Context(), api.LaunchParams{
				Template:  string(template),
				Env:       env,
				Namespace: namespace,
				Detach:    detach,
			}, os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to launch application: %w", err)
			}

			return nil
		},
	}

	c.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace to use")
	c.Flags().BoolP("detach", "d", false, "detach from the application")
	c.Flags().StringToStringVarP(&env, "env", "e", env, "set environment variables")
	c.Flags().StringVar(&envFile, "env-file", envFile, "set environment variables from a file")
	return c
}

func dockerComposeFile() string {
	if _, err := os.Stat("docker-compose.yml"); err == nil {
		return "docker-compose.yml"
	}

	if _, err := os.Stat("docker-compose.yaml"); err == nil {
		return "docker-compose.yaml"
	}

	return ""
}
