package completions

import (
	"context"
	"fmt"
	"strings"

	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
)

func PodOptions(client *api.Client, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	pods, err := getAside("pods", func() (*api.KraudPodList, error) {
		return client.ListPods(context.Background())
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var out []string
	for _, i := range pods.Items {
		out = append(out, i.Namespace+"/"+i.Name)
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}

func PodFromArg(ctx context.Context, client *api.Client, arg string) (string, error) {
	if !strings.Contains(arg, "/") {
		arg = "default/" + arg
	}

	ns, pod, ok := strings.Cut(arg, "/")
	if !ok {
		return arg, nil
	}

	pods, err := getAside("pods", func() (*api.KraudPodList, error) {
		return client.ListPods(context.Background())
	})
	if err != nil {
		return "", err
	}

	for _, i := range pods.Items {
		if i.Namespace == ns && i.Name == pod {
			return i.AID, nil
		}
	}

	return "", fmt.Errorf("pod %s not found", arg)
}
