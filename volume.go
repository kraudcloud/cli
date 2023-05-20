package main

import (
	"encoding/json"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"os"
)

func volumesCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "volume",
		Aliases: []string{"vol"},
		Short:   "manmage volumes",
	}

	c.AddCommand(volumeLs())

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
					table.AddRow(i.AID, i.Name, i.Class, i.Zone, humanize.Bytes(uint64(i.Size)))
				}
				table.Print()
			}
		},
	}

	return c
}
