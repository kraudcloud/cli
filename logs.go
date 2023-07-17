package main

import (
	"io"
	"os"

	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
)

func logsCMD() *cobra.Command {
	var follow bool

	c := &cobra.Command{
		Use:   "logs [CONTAINER]",
		Short: "logs of a container",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
				names = append(names, container.Namespace+"/"+container.Name)
			}

			return names, cobra.ShellCompDirectiveNoFileComp
		},
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
	c.Flags().BoolVar(&follow, "follow", false, "Keep tailing logs.")
	return c
}
