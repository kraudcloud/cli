package main

import (
	"fmt"

	"github.com/kraudcloud/cli/completions"
	"github.com/spf13/cobra"
	"strings"
)

func vpcsCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "vpc",
		Aliases: []string{"vpcs"},
		Short:   "manmage vpcs",
		Run:     vpcsLsRun,
	}

	c.AddCommand(&cobra.Command{
		Use:     "ls",
		Short:   "List vpcs",
		Aliases: []string{"list"},
		Args:    cobra.ExactArgs(0),
		Run:     vpcsLsRun,
	})
	c.AddCommand(vpcInspect())

	return c
}

func vpcsLsRun(cmd *cobra.Command, args []string) {
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
				fmt.Println("pods:")
				table = NewTable("id", "vpc", "overlay", "name")
				for _, i := range vv.Pods {
					overlays := []string{}
					for _, o := range i.Overlays {
						overlays = append(overlays, o.AID)
					}
					table.AddRow(i.ID, i.VpcIP, strings.Join(overlays, ","), i.Name+"."+i.Namespace)
				}
				table.Print()

				fmt.Println()
				fmt.Println("services:")
				table = NewTable("id", "vpc", "overlay", "name")
				for _, i := range vv.Services {
					table.AddRow(i.ID, i.VpcIP, strings.Join(i.OverlayRoutes, ","), i.Name+"."+i.Namespace)
				}
				table.Print()
			}
		},
	}

	return c
}
