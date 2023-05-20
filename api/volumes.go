package api

import (
	"context"
	"net/http"
)

func (c *Client) ListVolumes(ctx context.Context) (*KraudVolumeList, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/volumes",
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudVolumeList{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
