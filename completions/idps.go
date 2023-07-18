package completions

import (
	"context"
	"fmt"
	"strings"

	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
)

func IDPOptions(client *api.Client, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	idps, err := client.ListIDPs(context.Background())
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var out []string
	for _, i := range idps.Items {
		out = append(out, i.Namespace+"/"+i.Name)
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}

func IDPFromArg(ctx context.Context, client *api.Client, arg string) (string, error) {
	if !strings.Contains(arg, "/") {
		arg = "default/" + arg
	}

	ns, id, ok := strings.Cut(arg, "/")
	if !ok {
		return "", fmt.Errorf("invalid idp %s", arg)
	}

	idps, err := client.ListIDPs(ctx)
	if err != nil {
		return "", err
	}

	for _, i := range idps.Items {
		if i.Namespace == ns && i.Name == id {
			return *i.ID, nil
		}
	}

	return "", fmt.Errorf("pod %s not found", arg)
}
