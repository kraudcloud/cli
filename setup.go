package main

import (
	"bytes"
	"encoding/json"
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
				panic(err)
			}

			z, err := API().RotateUserCredentials(cmd.Context(), ".me", tokenName, "docker")
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

	c.Flags().StringVar(&tokenName, "token-name", "kra-setup-docker", "token name")

	return c
}

type dockerContext struct {
	Name string `json:"Name"`

	Endpoints struct {
		Docker struct {
			Host          string `json:"Host"`
			SkipTLSVerify bool   `json:"SkipTLSVerify"`
		} `json:"docker"`
	} `json:"Endpoints"`

	TLSMaterial struct {
		Docker []string `json:"docker"`
	} `json:"TLSMaterial"`

	Current bool `json:"Current"`

	Storage struct {
		MetadataPath string `json:"MetadataPath"`
		TLSPath      string `json:"TLSPath"`
	} `json:"Storage"`
}

func dockerContextInspect(context string) (dockerContext, error) {
	c := []dockerContext{}
	docker := exec.Command("docker", "context", "inspect", context)
	buf := bytes.NewBuffer(nil)
	docker.Stdout = buf
	docker.Stderr = os.Stderr

	err := docker.Run()
	if err != nil {
		return dockerContext{}, err
	}

	err = json.Unmarshal(buf.Bytes(), &c)
	if err != nil || len(c) == 0 {
		return dockerContext{}, err
	}

	return c[0], nil
}
