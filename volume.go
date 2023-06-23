package main

import (
	"encoding/json"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"os"
	"fmt"
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
				panic(err)
			}

			switch OUTPUT_FORMAT {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(vv)

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
		Run: func(cmd *cobra.Command, args []string) {
			err := API().DeleteVolume(cmd.Context(), args[0])
			if err != nil {
				panic(err)
			}
			fmt.Println("deleted")
		},
	}

	return c
}
