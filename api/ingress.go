package api

import (
	"context"
	"net/http"
)

func (c *Client) GetIngress(ctx context.Context, name string) (*K8sIngress, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/networking.k8s.io/v1/namespaces/default/ingresses/"+name,
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &K8sIngress{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
