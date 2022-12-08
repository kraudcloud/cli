package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/exp/maps"
)

func getConfigDir() string {
	type dockerConfig struct {
		CurrentContext string `json:"currentContext"`
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	ctx := os.Getenv("DOCKER_CONTEXT")
	if ctx == "" {
		f, err := os.Open(filepath.Join(home, ".docker", "config.json"))
		if err != nil {
			log.Fatalln(err)
		}

		var cfg dockerConfig
		if err := json.NewDecoder(f).Decode(&cfg); err != nil {
			log.Fatalln(err)
		}

		ctx = cfg.CurrentContext
	}

	return filepath.Join(home, "/.docker/contexts/tls", ctx)
}

func authClient() *http.Client {
	ca := filepath.Join(configDir, "ca.crt")
	cert := filepath.Join(configDir, "cert.pem")
	key := filepath.Join(configDir, "key.pem")

	if ca == "" {
		log.Fatal("No CA provided")
	}

	caCert, err := os.ReadFile(ca)
	if err != nil {
		log.Fatalln(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	if key == "" || cert == "" {
		log.Fatalln("key and cert must be provided")
	}

	pair, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Fatalln(err)
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{pair},
				RootCAs:            caCertPool,
				InsecureSkipVerify: true,
			},
		},
	}
}

func TableFromJSON(data []byte) (*tablewriter.Table, error) {
	var m []map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	if len(m) == 0 {
		return nil, fmt.Errorf("no data")
	}

	t := tablewriter.NewWriter(os.Stdout)
	t.SetBorder(false)
	t.SetHeaderLine(false)
	t.SetColumnSeparator("|")
	t.SetCenterSeparator(" ")
	t.SetAutoWrapText(false)

	keys := maps.Keys(m[0])
	t.SetHeader(keys)

	for _, v := range m {
		var row []string
		for _, k := range keys {
			s := fmt.Sprintf("%v", v[k])
			s = strings.ReplaceAll(s, "\n", " ")
			if len(s) > 36 {
				s = s[:33] + "..."
			}

			row = append(row, s)
		}
		t.Append(row)
	}

	return t, nil
}
