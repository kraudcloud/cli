package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func domainsCMD() *cobra.Command {

	c := &cobra.Command{
		Use:     "domains",
		Aliases: []string{"domain"},
		Short:   "Manage domains",
	}

	c.AddCommand(domainsLs())

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
