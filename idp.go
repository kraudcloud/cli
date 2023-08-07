package main

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"

	"github.com/kraudcloud/cli/completions"
	"github.com/spf13/cobra"
)

func idpCMD() *cobra.Command {

	c := &cobra.Command{
		Use:   "idp",
		Short: "Manage identity providers",
	}

	c.AddCommand(idpLs())
	c.AddCommand(idpGet())
	c.AddCommand(idpDelete())
	c.AddCommand(idpCreate())
	c.AddCommand(idpCert())

	return c
}

func idpCreate() *cobra.Command {

	var name string
	var namespace string
	var protocol string
	var metadataURL string

	c := &cobra.Command{
		Use:     "new",
		Short:   "Create an identity provider",
		Aliases: []string{"add", "new"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			u, err := url.Parse(metadataURL)
			if err != nil {
				return fmt.Errorf("error parsing url: %v", err)
			}

			meta, err := ReadURI(cmd.Context(), *u)
			if err != nil {
				return fmt.Errorf("error reading metadata: %v", err)
			}

			ig, err := API().CreateIDP(cmd.Context(),
				name,
				namespace,
				protocol,
				base64.StdEncoding.EncodeToString(meta),
			)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error creating identity provider: %v\n", err)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "identity provider %s created\n", ig.Name)
			return nil
		},
	}

	c.Flags().StringVar(&namespace, "namespace", "default", "Namespace to create resource in")
	c.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completions.NamespaceOptions(API(), cmd, args, toComplete)
	})

	c.Flags().StringVar(&name, "name", "", "Name of the identity provider")
	c.MarkFlagRequired("name")
	c.RegisterFlagCompletionFunc("name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	c.Flags().StringVar(&protocol, "proto", "saml", "Protocol of the identity provider")
	c.RegisterFlagCompletionFunc("proto", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	c.Flags().StringVar(&metadataURL, "svc", "", "service metadata url or local file")
	c.MarkFlagRequired("svc")

	return c
}

func idpLs() *cobra.Command {

	c := &cobra.Command{
		Use:   "ls",
		Short: "List identity providers",
		Run: func(cmd *cobra.Command, _ []string) {
			ig, err := API().ListIDPs(cmd.Context())
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error listing identity providers: %v\n", err)
				return
			}

			table := NewTable("ID", "Name", "Protocol")
			for _, idp := range ig.Items {
				table.AddRow(*idp.ID, idp.Namespace+"/"+idp.Name, idp.Protocol)
			}
			table.Print()
		},
	}

	return c
}

func idpGet() *cobra.Command {

	c := &cobra.Command{
		Use:     "inspect",
		Aliases: []string{"get"},
		Short:   "Inspect an identity provider",
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.IDPOptions(API(), cmd, args, toComplete)
		},
		Run: func(cmd *cobra.Command, args []string) {
			id := completions.IDPFromArg(cmd.Context(), API(), args[0])
			ig, err := API().InspectIDP(cmd.Context(), id)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error inspecting identity provider: %v\n", err)
				return
			}

			identJSONEncoder(cmd.OutOrStdout(), ig)
		},
	}

	return c
}

func idpCert() *cobra.Command {

	c := &cobra.Command{
		Use:   "cert",
		Short: "Get the certificate of an identity provider",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.IDPOptions(API(), cmd, args, toComplete)
		},
		Run: func(cmd *cobra.Command, args []string) {
			id := completions.IDPFromArg(cmd.Context(), API(), args[0])
			ig, err := API().InspectIDP(cmd.Context(), id)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error inspecting identity provider: %v\n", err)
				return
			}

			der, err := base64.StdEncoding.DecodeString(*ig.IdpCert)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error decoding certificate: %v\n", err)
				return
			}

			block := &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: der,
			}

			pem.Encode(os.Stdout, block)
		},
	}

	return c
}

func idpDelete() *cobra.Command {

	c := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm", "del", "remove"},
		Short:   "Delete an identity provider",
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completions.IDPOptions(API(), cmd, args, toComplete)
		},
		Run: func(cmd *cobra.Command, args []string) {
			id := completions.IDPFromArg(cmd.Context(), API(), args[0])
			err := API().DeleteIDP(cmd.Context(), id)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error deleting identity provider: %v\n", err)
				return
			}
		},
	}

	return c
}
