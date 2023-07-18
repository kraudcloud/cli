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

	pods, err := client.ListVolumes(context.Background())
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var out []string
	for _, i := range pods.Items {
		out = append(out, i.Name)
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}

func VolumeFromArg(ctx context.Context, client *api.Client, arg string) (string, error) {
	if !strings.Contains(arg, "/") {
		arg = "default/" + arg
	}

	volumes, err := client.ListVolumes(ctx)
	if err != nil {
		return "", err
	}

	for _, i := range volumes.Items {
		if i.Name == arg {
			return i.AID, nil
		}
	}

	return "", nil
}
