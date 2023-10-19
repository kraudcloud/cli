package api

import (
	"context"
	"net/http"
)

func (c *Client) ListVpcOverlays(ctx context.Context) (*KraudVpcOverlayList, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/vpc_overlays",
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudVpcOverlayList{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) GetVpcOverlay(ctx context.Context, q string) (*KraudVpcOverlay, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/vpc_overlays/"+q,
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudVpcOverlay{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
