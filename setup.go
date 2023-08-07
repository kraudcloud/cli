package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func setupCMD() *cobra.Command {
	c := &cobra.Command{
		Use:   "setup",
		Short: "Local setup",
	}

	c.AddCommand(dockerSetupCmd())

	return c
}

func dockerSetupCmd() *cobra.Command {

	var tokenName string

	c := &cobra.Command{
		Use:   "docker",
		Short: "setup local docker to access kraud via context",
		Run: func(cmd *cobra.Command, args []string) {

			me, err := API().GetUserMe(cmd.Context())
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error getting user: %v\n", err)
				return
			}

			z, err := API().RotateUserCredentials(cmd.Context(), ".me", tokenName, "docker")
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error rotating credentials: %v\n", err)
				return
			}

			docker := exec.Command("docker", "context", "import", "kraud."+me.Tenant.Org, "-")
			docker.Stdout = os.Stdout
			docker.Stderr = os.Stderr
			docker.Stdin = z

			err = docker.Run()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error importing context: %v\n", err)
				return
			}

			fmt.Println("")
			fmt.Println("to command your kraud cluster with docker, use")
			fmt.Println("   docker context use kraud." + me.Tenant.Org)
			fmt.Println("   docker info")
			fmt.Println("and to switch back to local docker")
			fmt.Println("   docker context use default")
			fmt.Println("alternatively prefix every command with --context kraud." + me.Tenant.Org)
			fmt.Println("   docker --context kraud." + me.Tenant.Org + " info")

		},
	}

	c.Flags().StringVar(&tokenName, "token-name", "kra-setup-docker", "token name")

	return c
}
