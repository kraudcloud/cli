package main

import (
	"fmt"

	"github.com/kraudcloud/cli/completions"
	"github.com/spf13/cobra"
)

func vpcOverlaysCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "overlays",
		Aliases: []string{"overlay", "vpc_overlays", "vpc_overlay"},
		Short:   "manage vpc overlays",
	}

	c.AddCommand(vpcOverlayLs())
	c.AddCommand(vpcOverlayInspect())

	return c
}

func vpcOverlayLs() *cobra.Command {

	c := &cobra.Command{
		Use:     "ls",
		Short:   "List vpcs",
		Aliases: []string{"list"},
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			vv, err := API().ListVpcOverlays(cmd.Context())
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error listing vpcs: %v\n", err)
				return
			}

			switch OUTPUT_FORMAT {
			case "json":
				identJSONEncoder(cmd.OutOrStdout(), vv)
			default:
				table := NewTable("aid", "namespace", "name", "driver", "net4", "net6")
				for _, i := range vv.Items {
					table.AddRow(i.AID, i.Namespace, i.Name, i.Driver, i.Net4, i.Net6)
				}
				table.Print()
			}
		},
	}

	return c
}

func vpcOverlayInspect() *cobra.Command {

	c := &cobra.Command{
		Use:     "show",
		Short:   "Inspect vpc",
		Aliases: []string{"get", "inspect"},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.VpcOptions(API(), cmd, args, toComplete)
		},

		Run: func(cmd *cobra.Command, args []string) {

			q := "default"
			if len(args) > 0 {
				q = args[0]
			}

			vv, err := API().GetVpcOverlay(cmd.Context(), q)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error getting vpc: %v\n", err)
				return
			}

			switch OUTPUT_FORMAT {
			case "json":
				identJSONEncoder(cmd.OutOrStdout(), vv)
			default:
				identJSONEncoder(cmd.OutOrStdout(), vv)

			}
		},
	}

	return c
}
