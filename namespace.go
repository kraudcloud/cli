package main

import (
	"fmt"
	"time"

	"github.com/kraudcloud/cli/completions"
	"github.com/spf13/cobra"
)

func namespacesCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "namespaces",
		Aliases: []string{"ns", "namespace"},
		Short:   "Manage namespaces",
	}

	c.AddCommand(
		namespacesListCMD(),
		namespacesDeleteCMD(),
		namespacesInspectCMD(),
		namespacesCreateCMD(),
	)

	return c
}

func namespacesListCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Short:   "List namespaces",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ns, err := API().ListNamespaces(cmd.Context())
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error listing namespaces: %v\n", err)
				return nil
			}

			t := NewTable("NAME", "CREATED")
			for _, n := range ns.Items {
				t.AddRow(*n.Metadata.Name, time.Time(*n.Metadata.CreationTimestamp).Format(time.DateTime))
			}

			t.Print()
			return nil
		},
	}

	return c
}

func namespacesDeleteCMD() *cobra.Command {
	force := false
	c := &cobra.Command{
		Use:     "delete <namespace>",
		Aliases: []string{"del", "rm"},
		Short:   "delete a namespace and all of its deployments",
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.NamespaceOptions(API(), cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := API().DeleteNamespace(cmd.Context(), args[0], force)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error deleting namespace: %v\n", err)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "namespace %q deleted\n", args[0])
			return nil
		},
	}

	c.Flags().BoolVar(&force, "force", false, "force delete")

	return c
}

func namespacesInspectCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "inspect <namespace>",
		Aliases: []string{"in", "show"},
		Short:   "Inspect a namespace",
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.NamespaceOptions(API(), cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ns, err := API().NamespaceOverview(cmd.Context(), args[0])
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error inspecting namespace: %v\n", err)
				return nil
			}

			return identJSONEncoder(cmd.OutOrStdout(), ns)
		},
	}

	return c
}

func namespacesCreateCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "create <namespace>",
		Aliases: []string{"new", "add", "mk"},
		Short:   "Create a namespace",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ns, err := API().CreateNamespace(cmd.Context(), args[0])
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error creating namespace: %v\n", err)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", *ns.ID)
			return nil
		},
	}

	return c
}
