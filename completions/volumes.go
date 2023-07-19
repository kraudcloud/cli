package completions

import (
	"context"
	"strings"

	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
)

func VolumeOptions(client *api.Client, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	volumes, err := getAside("volumes", client.AuthToken, func() (*api.KraudVolumeList, error) {
		return client.ListVolumes(cmd.Context())
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var out []string
	for _, i := range volumes.Items {
		out = append(out, i.Name)
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}

func VolumeFromArg(ctx context.Context, client *api.Client, arg string) string {
	newArg := arg
	if !strings.Contains(newArg, "/") {
		newArg = "default/" + newArg
	}

	volumes, err := getAside("volumes", client.AuthToken, func() (*api.KraudVolumeList, error) {
		return client.ListVolumes(ctx)
	})
	if err != nil {
		return arg
	}

	for _, i := range volumes.Items {
		if i.Name == newArg {
			return i.AID
		}
	}

	return arg
}
