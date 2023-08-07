package main

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

func layersCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "layers",
		Aliases: []string{"layer"},
		Short:   "Manage layers",
	}

	c.AddCommand(layersLs())

	return c
}

func layersLs() *cobra.Command {

	c := &cobra.Command{
		Use:   "ls",
		Short: "List remote layers",
		Run: func(cmd *cobra.Command, _ []string) {
			ls, err := API().ListLayers(cmd.Context())
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error listing layers: %v\n", err)
				return
			}

			table := NewTable("ID", "Size", "OciID", "Refcount", "Sha256")
			for _, i := range ls.Items {
				table.AddRow(
					i.ID,
					humanize.Bytes(uint64(i.Size)),
					i.OciID,
					i.Refcount,
					i.Sha256,
				)
			}
			table.Print()
		},
	}

	return c
}
