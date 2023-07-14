package main

import (
	"fmt"

	"github.com/kraudcloud/cli/api"
	"github.com/spf13/cobra"
)

func domainsCMD() *cobra.Command {

	c := &cobra.Command{
		Use:     "domains",
		Aliases: []string{"domain"},
		Short:   "Manage domains",
	}

	c.AddCommand(domainsLs())
	c.AddCommand(DomainsAdd())

	return c
}

func domainsLs() *cobra.Command {

	c := &cobra.Command{
		Use:   "ls",
		Short: "List domains",
		Run: func(cmd *cobra.Command, _ []string) {

			ig, err := API().GetIngress(cmd.Context(), "default")
			if err != nil {
				log.Fatalln(err)
			}

			routes := map[string]string{}

			for _, domain := range ig.Spec.Rules {
				if domain.Host != nil {
					routes[*domain.Host] = fmt.Sprintf("%d", len(domain.HTTP.Paths))
				}
			}

			table := NewTable("Domain", "Routes")
			for _, tls := range ig.Spec.TLS {
				for _, domain := range tls.Hosts {
					table.AddRow(
						domain,
						routes[domain],
					)
				}
			}
			table.Print()

		},
	}

	return c
}

func DomainsAdd() *cobra.Command {
	ingressID := ""
	c := &cobra.Command{
		Use:   "add <domain>",
		Short: "add domain",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			domain := args[0]

			_, err := API().CreateDomain(cmd.Context(), api.KraudDomainCreate{
				Name:      domain,
				IngressID: optNil(ingressID),
			})
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Println("Domain added")
		},
	}

	c.Flags().StringVar(&ingressID, "ingress", "", "ingress id")
	return c
}

// optNil returns nil if v is the zero value of T, otherwise returns a pointer to v.
func optNil[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}
