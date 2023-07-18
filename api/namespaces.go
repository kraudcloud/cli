package api

import (
	"context"
	"net/http"
)

func (c *Client) ListNamespaces(ctx context.Context) (*K8sNamespaceList, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/api/v1/namespaces",
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &K8sNamespaceList{}
	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
