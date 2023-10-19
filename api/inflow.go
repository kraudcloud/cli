package api

import (
	"context"
	"net/http"
)

func (c *Client) ListInflows(ctx context.Context) (*KraudInflowList, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/inflows",
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudInflowList{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) ListIngresses(ctx context.Context) (*KraudIngressList, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/ingresses",
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudIngressList{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
