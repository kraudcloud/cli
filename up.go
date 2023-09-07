package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kraudcloud/cli/api"
	"github.com/kraudcloud/cli/compose"
	"github.com/kraudcloud/cli/compose/envparser"
	"github.com/mitchellh/colorstring"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

			template, err := os.Open(file)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error reading docker-compose file: %v\n", err)
				return err
			}
			defer template.Close()

			// needed env neededVars
			neededVars, err := envparser.ParseTemplateVars(template)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error getting needed env vars: %v\n", err)
				return err
			}
			template.Seek(0, 0)

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

			apply, newTemplate, err := rewriteComposeLocal(template, namespace)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error rewriting local template: %s", err)
				return nil
			}

			err = apply(cmd.Context())
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error creating volumes to inject local files into: %s", err)
			}

			err = API().Launch(cmd.Context(), api.LaunchParams{
				Template:  newTemplate,
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

// rewriteComposeLocal takes in a compose file,
// it parses the volumes section.
// It generates an application function and a new spec
// that must be handled *after* applying.
func rewriteComposeLocal(r io.Reader, namespace string) (func(ctx context.Context) error, *bytes.Buffer, error) {
	f, err := compose.Parse(r)
	if err != nil {
		return nil, nil, err
	}

	apply, newF, err := f.Rewrite(namespace + "__local")
	if err != nil {
		return nil, nil, fmt.Errorf("error rewriting compose file from local paths: %w", err)
	}

	remarshalled := &bytes.Buffer{}
	err = yaml.NewEncoder(remarshalled).Encode(newF)
	if err != nil {
		return nil, nil, fmt.Errorf("error reincoding the file: %w", err)
	}

	return apply, remarshalled, nil
}
