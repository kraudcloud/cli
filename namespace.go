package main

import (
	"time"

	"github.com/kraudcloud/cli/completions"
	"github.com/spf13/cobra"
)

func namespacesCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "namespaces",
		Aliases: []string{"ns"},
		Short:   "Manage namespaces",
	}

	c.AddCommand(
		namespacesListCMD(),
		namespacesDeleteCMD(),
		namespacesInspectCMD(),
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
				return err
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
			return API().DeleteNamespace(cmd.Context(), args[0], force)
		},
	}

	c.Flags().BoolVar(&force, "force", false, "force delete")

	return c
}

func namespacesInspectCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "inspect <namespace>",
		Aliases: []string{"in"},
		Short:   "Inspect a namespace",
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.NamespaceOptions(API(), cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ns, err := API().NamespaceOverview(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			return indentJSONEncoder(cmd.OutOrStdout(), ns)
		},
	}

	return c
}
