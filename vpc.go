package main

import (
	"fmt"

	"github.com/kraudcloud/cli/completions"
	"github.com/spf13/cobra"
)

func vpcsCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "vpc",
		Aliases: []string{"vpcs"},
		Short:   "manmage vpcs",
	}

	c.AddCommand(vpcLs())
	c.AddCommand(vpcInspect())

	return c
}

func vpcLs() *cobra.Command {

	c := &cobra.Command{
		Use:     "ls",
		Short:   "List vpcs",
		Aliases: []string{"list"},
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			vv, err := API().ListVpcs(cmd.Context())
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error listing vpcs: %v\n", err)
				return
			}

			switch OUTPUT_FORMAT {
			case "json":
				identJSONEncoder(cmd.OutOrStdout(), vv)
			default:
				table := NewTable("aid", "name")
				for _, i := range vv.Items {
					table.AddRow(i.AID, i.Name)
				}
				table.Print()
			}
		},
	}

	return c
}

func vpcInspect() *cobra.Command {

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

			vv, err := API().GetVpc(cmd.Context(), q)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error getting vpc: %v\n", err)
				return
			}

			switch OUTPUT_FORMAT {
			case "json":
				identJSONEncoder(cmd.OutOrStdout(), vv)
			default:

				fmt.Println()
				fmt.Println("public networks:")
				table := NewTable("id", "segment", "net")
				for _, i := range vv.PublicNetworks {
					table.AddRow(i.ID, i.Segment, i.Net)
				}
				table.Print()

				fmt.Println()
				fmt.Println("services:")
				table = NewTable("id", "ip4", "ip6", "name")
				for _, i := range vv.Services {
					table.AddRow(i.ID, i.Ip4, i.Ip6, i.Name+"."+i.Namespace)
				}
				table.Print()

			}
		},
	}

	return c
}
