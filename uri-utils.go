package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func readHTTP(ctx context.Context, u url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	s, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer s.Body.Close()

	body, err := io.ReadAll(s.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func readFile(ctx context.Context, u url.URL) ([]byte, error) {
	if u.Path == "-" {
		return readStdin(ctx, u)
	}

	body, err := os.ReadFile(u.Path)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func readStdin(ctx context.Context, u url.URL) ([]byte, error) {
	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	return body, nil
}

var uriReaders = map[string]func(context.Context, url.URL) ([]byte, error){
	"http":  readHTTP,
	"https": readHTTP,
	"file":  readFile,
	"":      readFile,
}

func ReadURI(ctx context.Context, uri url.URL) ([]byte, error) {
	f, ok := uriReaders[uri.Scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported scheme: %s", uri.Scheme)
	}

	return f(ctx, uri)
}
