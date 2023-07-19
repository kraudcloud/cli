package completions

import (
	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
)

func NamespaceOptions(client *api.Client, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ns, err := getAside("namespaces", client.AuthToken, func() (*api.K8sNamespaceList, error) {
		return client.ListNamespaces(cmd.Context())
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var out []string
	for _, i := range ns.Items {
		if i.Metadata.Name == nil {
			continue
		}

		out = append(out, *i.Metadata.Name)
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}
