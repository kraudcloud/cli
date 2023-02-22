package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

func certs() *cobra.Command {
	c := &cobra.Command{
		Use:     "certs",
		Aliases: []string{"cert"},
		Short:   "Manage certificates",
	}

	c.AddCommand(certsGenerate())

	return c
}

func certsGenerate() *cobra.Command {
	namespace := "default"
	dns := []string{}
	format  := "table"

	c := &cobra.Command{
		Use:   "generate",
		Short: "Generate a certificate",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {

			u := url.URL{
				Path:   fmt.Sprintf("/apis/certificates.kraudcloud.com/v1/%s/generate/%s", url.PathEscape(namespace), url.PathEscape(args[0])),
				RawQuery: url.Values{
					"dns": dns,
				}.Encode(),
			}

			req, err := http.NewRequest(
				"POST",
				u.String(),
				nil,
			)
			if err != nil {
				log.Fatalln(err)
			}

			if format == "json" {
				req.Header.Set("Accept", "application/json")
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := Do(req)
			if err != nil {
				log.Fatalln(err)
			}

			defer resp.Body.Close()

			if resp.StatusCode > 201 {
				log.Warnln(resp.Status)
				var e errResp
				json.NewDecoder(resp.Body).Decode(&e)
				log.WithField("error", e.Error).WithField("message", e.Message).Fatalln("")
				os.Exit(resp.StatusCode)
			}

			io.Copy(os.Stdout, resp.Body)
		},
	}

	c.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace to create the certificate in")
	c.Flags().StringSliceVar(&dns, "dns", []string{}, "dns to add to the certificate")
	c.Flags().StringVarP(&format, "output", "o", "table", "output format (table|json)")

	return c
}
