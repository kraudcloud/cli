package completions

import (
	"context"
	"strings"

	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
)

func VpcOptions(client *api.Client, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	vpcs, err := client.ListVpcs(cmd.Context())
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var out []string
	for _, i := range vpcs.Items {
		out = append(out, i.Name)
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}

func VpcFromArg(ctx context.Context, client *api.Client, arg string) string {
	newArg := arg
	if !strings.Contains(newArg, "/") {
		newArg = "default/" + newArg
	}

	vpcs, err := client.ListVpcs(ctx)
	if err != nil {
		return arg
	}

	for _, i := range vpcs.Items {
		if i.Name == newArg {
			return i.AID
		}
	}

	return arg
}
