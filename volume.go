package main

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/kraudcloud/cli/completions"
	"github.com/spf13/cobra"
)

func volumesCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "volume",
		Aliases: []string{"vol"},
		Short:   "manmage volumes",
	}

	c.AddCommand(volumeLs())
	c.AddCommand(volumeRm())

	return c
}

func volumeLs() *cobra.Command {

	c := &cobra.Command{
		Use:     "ls",
		Short:   "List volumes",
		Aliases: []string{"list"},
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			vv, err := API().ListVolumes(cmd.Context())
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error listing volumes: %v\n", err)
				return
			}

			switch OUTPUT_FORMAT {
			case "json":
				identJSONEncoder(cmd.OutOrStdout(), vv)
			default:
				table := NewTable("aid", "name", "class", "zone", "size")
				for _, i := range vv.Items {
					zone := ""
					if i.Zone != nil {
						zone = *i.Zone
					}
					table.AddRow(i.AID, i.Name, i.Class, zone, humanize.Bytes(uint64(i.Size)))
				}
				table.Print()
			}
		},
	}

	return c
}

func volumeRm() *cobra.Command {

	c := &cobra.Command{
		Use:     "rm",
		Short:   "Remove volume",
		Aliases: []string{"remove", "del", "delete"},
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.VolumeOptions(API(), cmd, args, toComplete)
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := API().DeleteVolume(cmd.Context(), args[0])
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error deleting volume: %v\n", err)
				return
			}
			fmt.Println("deleted")
		},
	}

	return c
}
