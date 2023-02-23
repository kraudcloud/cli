package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) ListLayers(ctx context.Context) (*KraudLayerList, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/layers",
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudLayerList{}
	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) PushLayer(ctx context.Context, oid string, b io.Reader, size uint64, zsha string) (*KraudLayer, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("/apis/kraudcloud.com/v1/layers?size=%d&sha256=%s&oid=%s", size, zsha, oid),
		b,
	)

	req.Header.Set("Content-Type", "application/x-tar")

	if err != nil {
		return nil, err
	}

	var response = &KraudLayer{}
	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
