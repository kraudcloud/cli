package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("/apis/kraudcloud.com/v1/feeds/%s/apps", feedID),
				nil,
			)
			if err != nil {
				return err
			}

			partial := struct {
				Apps []any `json:"apps"`
			}{}

			err = API().Do(req, &partial)
			if err != nil {
				return err
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(partial)
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

			buf := &bytes.Buffer{}

			body := multipart.NewWriter(buf)
			file, err := body.CreateFormFile("template", "template.yaml")
			if err != nil {
				return err
			}

			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(file, f)
			if err != nil {
				return err
			}

			if changelog != "" {
				err = body.WriteField("changelog", changelog)
				if err != nil {
					return err
				}
			}

			err = body.Close()
			if err != nil {
				return err
			}

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("/apis/kraudcloud.com/v1/feeds/%s/app", feedID),
				buf,
			)
			if err != nil {
				return err
			}

			req.Header.Set("Content-Type", body.FormDataContentType())

			err = API().Do(req, nil)
			if err != nil {
				return err
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

			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("/apis/kraudcloud.com/v1/feeds/%s/apps/%s/template", feedID, appID),
				nil,
			)
			if err != nil {
				return err
			}

			var resp = map[string]interface{}{}

			err = API().Do(req, &resp)
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(resp)
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

func appsRun() *cobra.Command {
	feed := ""
	appArgs := map[string]string{}
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
					AdditionalProperties: appArgs,
				},
				ProjectName: namespace,
			}

			resp, err := API().LaunchApp(cmd.Context(), feedID, appID, body)
			if err != nil {
				return err
			}

			if ok, _ := cmd.Flags().GetBool("detach"); ok {
				return nil
			}

			return API().LaunchAttach(cmd.Context(), os.Stdout, resp.LaunchID)
		},
	}

	c.Flags().StringVarP(&feed, "feed", "f", "", "app store")
	c.MarkFlagRequired("feed")
	c.RegisterFlagCompletionFunc("feed", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completions.FeedOptions(API(), cmd, args, toComplete)
	})

	c.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace to run the app in")
	c.Flags().StringToStringVarP(&appArgs, "args", "a", appArgs, "arguments to pass to the app")
	c.Flags().BoolP("detach", "d", false, "detach from the app launch process")
	return c
}
