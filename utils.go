package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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
