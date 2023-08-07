package main

import (
	"fmt"
	"os"

	"github.com/kraudcloud/cli/api"
	"github.com/kraudcloud/cli/completions"
	"github.com/spf13/cobra"
)

func appsCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "apps",
		Aliases: []string{"apps"},
		Short:   "Manage apps",
	}

	c.AddCommand(appsPush())
	c.AddCommand(appsLs())
	c.AddCommand(appsInspect())
	c.AddCommand(appsRun())

	return c
}

func appsLs() *cobra.Command {
	feed := ""

	c := &cobra.Command{
		Use:     "ls",
		Short:   "List apps",
		Aliases: []string{"l"},
		RunE: func(cmd *cobra.Command, args []string) error {
			feedID := completions.FeedFromArg(cmd.Context(), API(), feed)

			apps, err := API().ListApps(cmd.Context(), feedID)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error listing apps: %v\n", err)
				return nil
			}

			return identJSONEncoder(cmd.OutOrStdout(), apps)
		},
	}

	c.Flags().StringVarP(&feed, "feed", "f", "", "app store")
	c.MarkFlagRequired("feed")
	c.RegisterFlagCompletionFunc("feed", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completions.FeedOptions(API(), cmd, args, toComplete)
	})

	return c
}

func appsPush() *cobra.Command {
	feed := ""

	changelog := ""
	c := &cobra.Command{
		Use:     "push <app.yaml>",
		Short:   "Push an app to the kraud",
		Aliases: []string{"p"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			feedID := completions.FeedFromArg(cmd.Context(), API(), feed)

			template, err := os.Open(args[0])
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error opening app.yaml: %v\n", err)
				return nil
			}
			defer template.Close()

			err = API().PushApp(cmd.Context(), feedID, template, changelog)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error pushing app: %v\n", err)
				return nil
			}

			return nil
		},
	}

	c.Flags().StringVarP(&feed, "feed", "f", "", "app store")
	c.Flags().StringVar(&changelog, "changelog", "", "changelog for the app")
	c.MarkFlagRequired("feed")
	c.RegisterFlagCompletionFunc("feed", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completions.FeedOptions(API(), cmd, args, toComplete)
	})

	return c
}

func appsInspect() *cobra.Command {
	feed := ""

	c := &cobra.Command{
		Use:     "inspect",
		Short:   "Inspect an app",
		Aliases: []string{"i"},
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// feed must exist to complete apps
			if feed == "" {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return completions.AppOptions(API(), feed, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			feedID := completions.FeedFromArg(cmd.Context(), API(), feed)
			appID := completions.AppFromArg(cmd.Context(), API(), feedID, args[0])

			app, err := API().InspectApp(cmd.Context(), feedID, appID)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error inspecting app: %v\n", err)
				return nil
			}

			return identJSONEncoder(cmd.OutOrStdout(), app)
		},
	}

	c.Flags().StringVarP(&feed, "feed", "f", "", "store to push to")
	c.MarkFlagRequired("feed")
	c.RegisterFlagCompletionFunc("feed", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completions.FeedOptions(API(), cmd, args, toComplete)
	})

	return c
}

type errResp struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// TODO: maybe support .env files?
func appsRun() *cobra.Command {
	feed := ""
	env := map[string]string{}
	namespace := "default"

	c := &cobra.Command{
		Use:    "run <app>",
		Hidden: true, // hide launch apps for now. needs polish
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// feed must exist to complete apps
			if feed == "" {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return completions.AppOptions(API(), feed, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			feedID := completions.FeedFromArg(cmd.Context(), API(), feed)
			appID := completions.AppFromArg(cmd.Context(), API(), feedID, args[0])

			body := api.KraudLaunchSettings{
				Config: api.KraudLaunchSettings_Config{
					AdditionalProperties: env,
				},
				ProjectName: namespace,
			}

			resp, err := API().LaunchApp(cmd.Context(), feedID, appID, body)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error launching app: %v\n", err)
				return nil
			}

			if ok, _ := cmd.Flags().GetBool("detach"); ok {
				return nil
			}

			err = API().LaunchAttach(cmd.Context(), os.Stdout, resp.LaunchID)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error attaching to app: %v\n", err)
				return nil
			}

			return nil
		},
	}

	c.Flags().StringVarP(&feed, "feed", "f", "", "app store")
	c.MarkFlagRequired("feed")
	c.RegisterFlagCompletionFunc("feed", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completions.FeedOptions(API(), cmd, args, toComplete)
	})

	c.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace to run the app in")
	c.Flags().StringToStringVarP(&env, "env", "e", env, "environment to pass to the app")
	c.Flags().BoolP("detach", "d", false, "detach from the app launch process")
	return c
}
