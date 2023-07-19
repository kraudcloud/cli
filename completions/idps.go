package completions

import (
	"context"
	"strings"

	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
)

func IDPOptions(client *api.Client, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	idps, err := getAside("idps", client.AuthToken, func() (*api.KraudIdentityProviderList, error) {
		return client.ListIDPs(context.Background())
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var out []string
	for _, i := range idps.Items {
		out = append(out, i.Namespace+"/"+i.Name)
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}

func IDPFromArg(ctx context.Context, client *api.Client, arg string) string {
	newArg := arg
	if !strings.Contains(newArg, "/") {
		newArg = "default/" + newArg
	}

	ns, id, ok := strings.Cut(newArg, "/")
	if !ok {
		return arg
	}

	idps, err := getAside("idps", client.AuthToken, func() (*api.KraudIdentityProviderList, error) {
		return client.ListIDPs(ctx)
	})
	if err != nil {
		return arg
	}

	for _, i := range idps.Items {
		if i.Namespace == ns && i.Name == id {
			return *i.ID
		}
	}

	return arg
}
