package main

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	var svc_metadata_url string

	c := &cobra.Command{
		Use:     "new",
		Short:   "Create an identity provider",
		Aliases: []string{"add", "new"},
		Run: func(cmd *cobra.Command, _ []string) {

			metadata := ""

			u, err := url.Parse(svc_metadata_url)
			if err != nil {
				log.Fatalln(err)
			}

			if u.Scheme == "http" || u.Scheme == "https" {
				resp, err := http.Get(u.String())
				if err != nil {
					log.Fatalln(err)
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatalln(err)
				}
				metadata = string(body)
			} else {

				f, err := os.Open(u.Path)
				if err != nil {
					log.Fatalln(err)
				}
				defer f.Close()

				body, err := ioutil.ReadAll(f)
				if err != nil {
					log.Fatalln(err)
				}

				metadata = string(body)
			}

			b64metadata := base64.StdEncoding.EncodeToString([]byte(metadata))

			ig, err := API().CreateIDP(cmd.Context(),
				name,
				namespace,
				protocol,
				b64metadata,
			)
			if err != nil {
				log.Fatalln(err)
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(ig)
		},
	}

	c.Flags().StringVar(&namespace, "namespace", "default", "Namespace to create resource in")

	c.Flags().StringVar(&name, "name", "", "Name of the identity provider")
	c.MarkFlagRequired("name")
	c.Flags().StringVar(&protocol, "proto", "saml", "Protocol of the identity provider")
	c.Flags().StringVar(&svc_metadata_url, "svc", "", "service metadata url or local file")
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
				log.Fatalln(err)
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
		Run: func(cmd *cobra.Command, args []string) {

			ig, err := API().InspectIDP(cmd.Context(), args[0])
			if err != nil {
				log.Fatalln(err)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(ig)

		},
	}

	return c
}

func idpCert() *cobra.Command {

	c := &cobra.Command{
		Use:   "cert",
		Short: "Get the certificate of an identity provider",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			ig, err := API().InspectIDP(cmd.Context(), args[0])
			if err != nil {
				log.Fatalln(err)
			}

			der, err := base64.StdEncoding.DecodeString(*ig.IdpCert)
			if err != nil {
				log.Fatalln(err)
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
		Run: func(cmd *cobra.Command, args []string) {

			err := API().DeleteIDP(cmd.Context(), args[0])
			if err != nil {
				log.Fatalln(err)
			}

		},
	}

	return c
}
