package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func inflowsCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "in",
		Aliases: []string{"inflows", "inflow"},
		Short:   "manage inflows",
		Args:    cobra.ExactArgs(0),
		Run:     runInflowsList,
	}

	c.AddCommand(&cobra.Command{
		Use:     "flow",
		Short:   "List inflows",
		Aliases: []string{"list", "ls"},
		Args:    cobra.ExactArgs(0),
		Run:     runInflowsList,
	})

	c.AddCommand(&cobra.Command{
		Use:     "ig",
		Short:   "List ingresses",
		Aliases: []string{"ingresses", "ingress"},
		Args:    cobra.ExactArgs(0),
		Run:     runIngressList,
	})

	return c
}

func runInflowsList(cmd *cobra.Command, args []string) {
	vv, err := API().ListInflows(cmd.Context())
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "error listing inflows: %v\n", err)
		return
	}

	switch OUTPUT_FORMAT {
	case "json":
		identJSONEncoder(cmd.OutOrStdout(), vv)
	default:
		table := NewTable("vpc", "kind", "public", "target")
		for _, i := range vv.Items {
			table.AddRow(i.VpcName, i.Kind, i.DisplayPublic, i.DisplayTarget)
		}
		table.Print()
	}
}

func runIngressList(cmd *cobra.Command, args []string) {
	vv, err := API().ListIngresses(cmd.Context())
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "error listing ingresses: %v\n", err)
		return
	}

	switch OUTPUT_FORMAT {
	case "json":
		identJSONEncoder(cmd.OutOrStdout(), vv)
	default:
		table := NewTable("id", "domain")
		for _, i := range vv.Items {
			table.AddRow(i.ID, i.IngressDomain)
		}
		table.Print()
	}
}
