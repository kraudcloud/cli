package main

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
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
	pods, err := API().ListPods(cmd.Context())
	if err != nil {
		panic(err)
	}

	switch OUTPUT_FORMAT {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(pods)

	default:
		table := NewTable("aid", "namespace", "name", "cpu", "mem", "image")
		for _, i := range pods.Items {

			var image string
			if len(i.Containers) > 0 {
				image = i.Containers[0].ImageName
			}

			if strings.HasPrefix(image, "index.docker.io/library/") {
				image = strings.TrimPrefix(image, "index.docker.io/library/")
			}

			if len(i.Namespace) > 20 {
				i.Namespace = i.Namespace[:18] + ".."
			}

			table.AddRow(i.AID, i.Namespace, i.Name,
				i.CPU,
				i.Mem,
				image)
		}
		table.Print()
	}
}

func podsInspect() *cobra.Command {
	c := &cobra.Command{
		Use:               "inspect",
		Short:             "Inspect pod",
		Aliases:           []string{"get", "show", "info", "i"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: podAIDComplete,
		Run: func(cmd *cobra.Command, args []string) {

			pod, err := API().InspectPod(cmd.Context(), args[0])
			if err != nil {
				panic(err)
			}

			switch OUTPUT_FORMAT {
			default:
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(pod)
			}
		},
	}
	return c
}

func podsEdit() *cobra.Command {

	c := &cobra.Command{
		Use:               "edit",
		Short:             "Edit pod",
		Aliases:           []string{"e"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: podAIDComplete,
		Run: func(cmd *cobra.Command, args []string) {

			pod, err := API().InspectPod(cmd.Context(), args[0])
			if err != nil {
				panic(err)
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}

			tmpfile, err := ioutil.TempFile("", "kra-pod-")
			if err != nil {
				panic(err)
			}
			defer os.Remove(tmpfile.Name())

			pod.ID = nil
			for i := range pod.Containers {
				pod.Containers[i].ID = nil
			}

			str, err := json.MarshalIndent(pod, "", "  ")
			if err != nil {
				panic(err)
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
					panic(err)
				}

				strNu, err := os.ReadFile(tmpfile.Name())
				if err != nil {
					panic(err)
				}

				if string(str) == string(strNu) {
					log.Info("No changes")
					return
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

		},
	}
	return c
}

func podLogs() *cobra.Command {
	var follow bool

	c := &cobra.Command{
		Use:               "logs [CONTAINER]",
		Short:             "logs of a container",
		Aliases:           []string{"log"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: podAIDComplete,
		RunE: func(cmd *cobra.Command, args []string) error {
			rsp, err := API().GetLogs(cmd.Context(), args[0], api.LogsOptions{
				Follow: follow,
			})
			if err != nil {
				return err
			}
			defer rsp.Close()

			_, err = io.Copy(os.Stdout, rsp)
			return err
		},
	}

	// TODO: add --since, --tail, --timestamps when implemented
	c.Flags().BoolVarP(&follow, "follow", "f", false, "Keep tailing logs.")
	return c

}

func podAIDComplete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// command takes 1 arg
	if (len(args) > 1) || (len(args) == 1 && args[0] != "") {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Likely need to cache this?
	pods, err := API().ListPods(cmd.Context())
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var names []string
	for _, container := range pods.Items {
		names = append(names, container.AID)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
