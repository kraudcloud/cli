package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/kraudcloud/cli/api"
	"github.com/kraudcloud/cli/completions"
	"github.com/spf13/cobra"
)

func volumesCMD() *cobra.Command {
	c := &cobra.Command{
		Use:     "volume",
		Aliases: []string{"vol"},
		Args:    cobra.ExactArgs(0),
		Short:   "manage volumes",
		Run:     runVolumeLs,
	}

	c.AddCommand(&cobra.Command{
		Use:     "ls",
		Short:   "List volumes",
		Aliases: []string{"list"},
		Args:    cobra.ExactArgs(0),
		Run:     runVolumeLs,
	})

	c.AddCommand(volumeRm())
	c.AddCommand(volumeCreate())
	c.AddCommand(volumeCopy())

	return c
}

func runVolumeLs(cmd *cobra.Command, args []string) {
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

func volumeCreate() *cobra.Command {
	driver := ""
	size := ""
	additionalLabels := map[string]string{}
	volOpts := map[string]string{}

	c := &cobra.Command{
		Use:   "create",
		Short: "Create a volume with the specified name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			r := api.DockerVolumeCreateJSONRequestBody{
				Name: &name,
			}

			if driver != "" {
				r.Driver = &driver
			}

			if size != "" {
				volOpts["size"] = size
			}

			if len(volOpts) > 0 {
				r.DriverOpts = &api.VolumeCreateOptions_DriverOpts{
					AdditionalProperties: volOpts,
				}
			}

			if len(additionalLabels) > 0 {
				r.Labels = &api.VolumeCreateOptions_Labels{
					AdditionalProperties: additionalLabels,
				}
			}

			_, err := API().CreateVolume(cmd.Context(), r)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error creating volume: %v\n", err)
			}
		},
	}

	c.Flags().StringVarP(&driver, "driver", "d", driver, "volume driver")
	c.Flags().StringVarP(&size, "size", "s", size, "volume size")
	c.Flags().StringToStringVarP(&additionalLabels, "label", "l", additionalLabels, "volume labels")
	c.Flags().StringToStringVar(&volOpts, "opt", volOpts, "volume options")

	return c
}

func volumeCopy() *cobra.Command {
	namespace := "default"
	c := &cobra.Command{
		Use:   "copy src:dst",
		Short: "copy local dir/file to remote volume",
		Long: `copy local dir/file to remote volume,
src and dst are in format: local_path:volume_path.

Prefer using docker cp instead of this command, for both performance and reliability reasons.

This is here only to help with debugging "kra up".`,
		Args:   cobra.ExactArgs(1),
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			local, volume, ok := strings.Cut(args[0], ":")
			if !ok {
				return errors.New("invalid format")
			}

			volume = filepath.Clean(volume)
			local = filepath.Clean(local)

			err := API().UploadDir(cmd.Context(), namespace, local, volume)
			if err != nil {
				volName, rest, _ := strings.Cut(volume, string(os.PathSeparator))
				if rest == "" {
					rest = "/"
				}

				fmt.Fprintf(cmd.ErrOrStderr(), "error copying volume (namespace=%s volume=%s path=%s): %v\n", namespace, volName, rest, err)
				return nil
			}
			fmt.Println("copied sucessfully")

			return nil
		},
	}

	c.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace")
	return c
}
