package completions

import (
	"context"

	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
)

func FeedOptions(client *api.Client, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	feeds, err := client.ListFeeds(cmd.Context())
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var out []string
	for _, i := range feeds {
		out = append(out, i.Name)
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}

func FeedFromArg(ctx context.Context, client *api.Client, arg string) string {
	feeds, err := client.ListFeeds(ctx)
	if err != nil {
		return arg
	}

	for _, i := range feeds {
		if i.Name == arg {
			return i.ID
		}
	}

	return arg
}

func AppOptions(client *api.Client, feed string, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	feedID := FeedFromArg(cmd.Context(), client, feed)
	apps, err := client.ListApps(cmd.Context(), feedID)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var out []string
	for _, i := range apps.Items {
		out = append(out, i.Name)
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}

func AppFromArg(ctx context.Context, client *api.Client, feedID, arg string) string {
	apps, err := client.ListApps(ctx, feedID)
	if err != nil {
		return arg
	}

	for _, i := range apps.Items {
		if i.Name == arg {
			return i.ID
		}
	}

	return arg
}
