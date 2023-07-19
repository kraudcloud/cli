package completions

import (
	"context"
	"fmt"

	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
)

func FeedOptions(client *api.Client, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	feeds, err := getAside("feeds", client.AuthToken, func() (api.KraudFeedList, error) {
		return client.ListFeeds(cmd.Context())
	})
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
	feeds, err := getAside("feeds", client.AuthToken, func() (api.KraudFeedList, error) {
		return client.ListFeeds(ctx)
	})
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
	apps, err := getAside(fmt.Sprintf("%s:apps", feedID), client.AuthToken, func() (*api.ListAppsResponse, error) {
		return client.ListApps(cmd.Context(), feedID)
	})
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
	apps, err := getAside(fmt.Sprintf("%s:apps", feedID), client.AuthToken, func() (*api.ListAppsResponse, error) {
		return client.ListApps(ctx, feedID)
	})
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
