package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
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

	c := &cobra.Command{
		Use:   "docker",
		Short: "setup local docker to access kraud via context",
		Run: func(cmd *cobra.Command, args []string) {

			hostname, err := os.Hostname()
			if err != nil {
				panic(err)
			}

			me, err := API().GetUserMe(cmd.Context())
			if err != nil {
				panic(err)
			}

			z, err := API().RotateUserCredentials(cmd.Context(), ".me", hostname+"-docker-via-kra", "docker")
			if err != nil {
				panic(err)
			}

			docker := exec.Command("docker", "context", "import", "kraud."+me.Tenant.Org, "-")
			docker.Stdout = os.Stdout
			docker.Stderr = os.Stderr
			docker.Stdin = z

			err = docker.Run()
			if err != nil {
				panic(err)
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

	return c
}
