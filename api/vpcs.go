package api

import (
	"context"
	"net/http"
)

func (c *Client) ListVpcs(ctx context.Context) (*KraudVpcList, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/vpcs",
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudVpcList{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) GetVpc(ctx context.Context, q string) (*KraudVpc, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/vpcs/"+q,
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudVpc{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
