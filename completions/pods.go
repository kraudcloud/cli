package completions

import (
	"context"
	"strings"

	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
)

func PodOptions(client *api.Client, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	pods, err := client.ListPods(cmd.Context(), false)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var out []string
	for _, i := range pods.Items {
		out = append(out, i.Namespace+"/"+i.Name)
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}

func PodFromArg(ctx context.Context, client *api.Client, arg string) string {
	newArg := arg
	if !strings.Contains(newArg, "/") {
		newArg = "default/" + newArg
	}

	ns, pod, ok := strings.Cut(newArg, "/")
	if !ok {
		return arg
	}

	pods, err := client.ListPods(ctx, false)
	if err != nil {
		return arg
	}

	for _, i := range pods.Items {
		if i.Namespace == ns && i.Name == pod {
			return i.AID
		}
	}

	return arg
}
