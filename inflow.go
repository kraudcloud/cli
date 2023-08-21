package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func inflowsCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "in",
		Aliases: []string{"inflows"},
		Short:   "manage inflows",
		Run:     runInflowsList,
	}

	c.AddCommand(&cobra.Command{
		Use:     "ls",
		Short:   "List inflows",
		Aliases: []string{"list"},
		Args:    cobra.ExactArgs(0),
		Run:     runInflowsList,
	})

	return c
}

func runInflowsList(cmd *cobra.Command, args []string) {
	vv, err := API().ListInflows(cmd.Context())
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "error listing vpcs: %v\n", err)
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
