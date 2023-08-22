package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/fatih/color"
	"github.com/kraudcloud/cli/api"
	"github.com/kraudcloud/cli/completions"
	"github.com/kraudcloud/cli/compose/envparser"
	"github.com/mattn/go-tty"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

func psCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "ps",
		Aliases: []string{"pod"},
		Short:   "List pods",
		Run:     podsLsRun,
	}

	c.AddCommand(podsLs())

	return c
}
func podsCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "pods",
		Aliases: []string{"pod"},
		Short:   "Manage pods",
	}

	c.AddCommand(podsLs())
	c.AddCommand(podsInspect())
	c.AddCommand(podsEdit())
	c.AddCommand(podLogs())
	c.AddCommand(podSSH())

	return c
}

func podsLs() *cobra.Command {

	c := &cobra.Command{
		Use:     "ls",
		Short:   "List pods",
		Aliases: []string{"list", "ps"},
		Run:     podsLsRun,
	}

	return c
}

func podsLsRun(cmd *cobra.Command, args []string) {
	pods, err := API().ListPods(cmd.Context(), true)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "error getting pods: %v\n", err)
		return
	}

	switch OUTPUT_FORMAT {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(pods)

	default:
		table := NewTable("aid", "namespace", "name", "cpu", "mem", "status", "image")
		for _, i := range pods.Items {

			var image string
			if len(i.Containers) > 0 {
				image = i.Containers[0].ImageName
			}

			image = strings.TrimPrefix(image, "index.docker.io/library/")
			if len(i.Namespace) > 20 {
				i.Namespace = i.Namespace[:18] + ".."
			}

			var status = "?"
			if i.Status != nil {
				status = i.Status.Display
				if len(status) > len("Terminated") {
					status = strings.Split(status, " ")[0]
				}
				if len(status) > len("Terminated") {
					status = status[:len("Terminated")]
				}
				if i.Status.Healthy {
					status = color.GreenString(status)
				} else {
					status = color.RedString(status)
				}
			}

			if len(image) > 24 {
				ss := strings.Split(image, "/")
				if len(ss) > 1 {
					image = ss[len(ss)-1]
				}
			}

			if len(image) > 24 {
				image = image[:23] + ".."
			}

			table.AddRow(i.AID, i.Namespace, i.Name,
				i.CPU,
				i.Mem,
				status,
				image,
			)
		}
		table.Print()
	}
}

func podsInspect() *cobra.Command {
	c := &cobra.Command{
		Use:     "inspect",
		Short:   "Inspect pod",
		Aliases: []string{"get", "show", "info", "i"},
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.PodOptions(API(), cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			aid := completions.PodFromArg(cmd.Context(), API(), args[0])
			pod, err := API().InspectPod(cmd.Context(), aid)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error getting pod: %v\n", err)
				return nil
			}

			return identJSONEncoder(cmd.OutOrStdout(), pod)
		},
	}
	return c
}

func podsEdit() *cobra.Command {
	c := &cobra.Command{
		Use:     "edit",
		Short:   "Edit pod",
		Aliases: []string{"e"},
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.PodOptions(API(), cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			aid := completions.PodFromArg(cmd.Context(), API(), args[0])
			pod, err := API().InspectPod(cmd.Context(), aid)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error getting pod: %v\n", err)
				return nil
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}

			tmpfile, err := os.CreateTemp("", "kra-pod-")
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error creating temp file: %v\n", err)
				return nil
			}
			defer os.Remove(tmpfile.Name())

			pod.ID = nil
			for i := range pod.Containers {
				pod.Containers[i].ID = nil
			}

			str, err := json.MarshalIndent(pod, "", "  ")
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error marshalling pod: %v\n", err)
				return nil
			}

			tmpfile.Write(str)
			tmpfile.Close()

			for {

				edit := exec.Command(editor, tmpfile.Name())
				edit.Stdin = os.Stdin
				edit.Stdout = os.Stdout
				edit.Stderr = os.Stderr
				err = edit.Run()

				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "error running editor: %v\n", err)
					return nil
				}

				strNu, err := os.ReadFile(tmpfile.Name())
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "error reading file: %v\n", err)
					return nil
				}

				if string(str) == string(strNu) {
					log.Info("No changes")
					return nil
				}

				var newPod api.KraudPod
				err = json.Unmarshal(strNu, &newPod)
				if err != nil {

					log.Error("Error parsing json: ", err)
					log.Error("Press enter to go back to editor or ctrl+c to exit")
					bufio.NewReader(os.Stdin).ReadBytes('\n')

					continue
				}

				err = API().EditPod(cmd.Context(), pod.AID, &newPod)
				if err != nil {

					log.Error("changes rejected: ", err)
					log.Error("Press enter to go back to editor or ctrl+c to exit")
					bufio.NewReader(os.Stdin).ReadBytes('\n')

					continue
				}

				break
			}

			log.Info("changes commited but will not be applied until pod is restarted")
			return nil
		},
	}
	return c
}

func podLogs() *cobra.Command {
	var follow bool

	c := &cobra.Command{
		Use:     "logs [CONTAINER]",
		Short:   "logs of a container",
		Aliases: []string{"log"},
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.PodOptions(API(), cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			dockerClient := API().DockerClient()
			aid := completions.PodFromArg(ctx, API(), args[0])
			c, err := dockerClient.ContainerInspect(ctx, aid)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error getting container: %v\n", err)
				return nil
			}

			options := types.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Follow:     follow,
				Tail:       "all",
			}
			responseBody, err := dockerClient.ContainerLogs(ctx, c.ID, options)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error getting logs: %v\n", err)
				return nil
			}
			defer responseBody.Close()

			if c.Config.Tty {
				_, err = io.Copy(os.Stdout, responseBody)
			} else {
				_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, responseBody)
			}

			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error copying logs: %v\n", err)
				return nil
			}

			return nil
		},
	}

	// TODO: add --since, --tail, --timestamps when implemented
	c.Flags().BoolVarP(&follow, "follow", "f", false, "Keep tailing logs.")
	return c

}

func podSSH() *cobra.Command {
	env := map[string]string{}
	envFile := ""
	user := ""
	workdir := ""
	envF := []string{}

	c := &cobra.Command{
		Use:   "ssh [CONTAINER]",
		Short: "ssh into a container",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.PodOptions(API(), cmd, args, toComplete)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			out := map[string]string{}
			if envFile != "" {
				f, err := os.Open(envFile)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "error opening env file: %v\n", err)
					return
				}
				defer f.Close()

				maps.Copy(out, envparser.EnvMapFromReader(f))
			}

			maps.Copy(out, env)

			for k, v := range out {
				envF = append(envF, fmt.Sprintf("%s=%s", k, v))
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			tty, err := tty.Open()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "stdin is not a terminal (%v)\n", err)
				return nil
			}

			err = API().SSH(ctx, tty, api.SSHParams{
				PodID:   completions.PodFromArg(ctx, API(), args[0]),
				User:    user,
				WorkDir: workdir,
				Env:     envF,
			})
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error getting container: %v\n", err)
				return nil
			}

			return nil
		},
	}

	c.Flags().StringToStringVarP(&env, "env", "e", env, "Set environment variables")
	c.Flags().StringVar(&envFile, "env-file", envFile, "Read in a file of environment variables")
	c.Flags().StringVarP(&user, "user", "u", user, "Username to use when connecting to the container")
	c.Flags().StringVarP(&workdir, "workdir", "w", workdir, "Working directory for the container")

	return c
}
