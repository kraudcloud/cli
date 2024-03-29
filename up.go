package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kraudcloud/cli/api"
	"github.com/kraudcloud/cli/compose/envparser"
	"github.com/mitchellh/colorstring"
	"github.com/spf13/cobra"
)

func UpCMD() *cobra.Command {
	cwd, _ := os.Getwd()
	cwd = filepath.Base(cwd)
	file := dockerComposeFile()

	namespace := cwd
	if namespace == "." {
		namespace = ""
	}

	env := map[string]string{}
	envFile := ".env"
	verbose := 0

	c := &cobra.Command{
		Use:   "up",
		Short: "run an application",
		RunE: func(cmd *cobra.Command, args []string) error {
			if namespace == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "namespace is required\n")
				return nil
			}

			template, err := os.ReadFile(file)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error reading docker-compose file: %v\n", err)
				return err
			}

			// needed env neededVars
			neededVars, err := envparser.ParseTemplateVars(bytes.NewReader(template))
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error getting needed env vars: %v\n", err)
				return err
			}

			loaders := []envparser.EnvLoader{
				envparser.LoadKV(env),
				envparser.LoadKVs(os.Environ()),
			}

			// load env vars from file
			if envFile != "" {
				f, err := os.Open(envFile)
				switch {
				case os.IsNotExist(err):
					break
				case err != nil:
					fmt.Fprintf(cmd.ErrOrStderr(), "error reading env file: %v\n", err)
					return nil
				}

				defer f.Close()

				loaders = append(loaders, envparser.LoadEnvReader(f))
			}

			// load env vars
			env, err := envparser.LoadEnv(neededVars, loaders...)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error loading env vars: %v\n", err)
				return nil
			}

			if verbose > 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "looking for env vars:\n")
				for k, v := range neededVars {
					fmt.Fprintf(cmd.ErrOrStderr(), "  %s", k)
					if v.Default != "" {
						fmt.Fprintf(cmd.ErrOrStderr(), "  (default: %s)", v.Default)
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "\n")
				}

				fmt.Fprintf(cmd.ErrOrStderr(), "env vars:\n")
				for k, v := range env {
					fmt.Fprintf(cmd.ErrOrStderr(), "  %s=%s\n", k, v)
				}
			}

			detach, _ := cmd.Flags().GetBool("detach")

			err = API().Launch(cmd.Context(), api.LaunchParams{
				Template:  string(template),
				Env:       env,
				Namespace: namespace,
				Detach:    detach,
			}, cmd.OutOrStdout())
			if err != nil {
				colorstring.Fprintf(cmd.ErrOrStderr(), "[red]%v\n", err)
				os.Exit(1)
				return nil
			}

			return nil
		},
	}

	c.Flags().StringVarP(&file, "file", "f", file, "docker-compose file to use")
	c.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace to use")
	c.Flags().BoolP("detach", "d", false, "detach from the application")
	c.Flags().StringToStringVarP(&env, "env", "e", env, "set environment variables")
	c.Flags().StringVar(&envFile, "env-file", envFile, "set environment variables from a file")
	c.Flags().CountVarP(&verbose, "verbose", "v", "verbose output")
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
