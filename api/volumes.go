package api

import (
	"bytes"
	"context"
	"encoding/json"
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

func (c *Client) DeleteVolume(ctx context.Context, aid string) error {

	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		"/apis/kraudcloud.com/v1/volumes/"+aid,
		nil,
	)

	if err != nil {
		return err
	}

	return c.Do(req, nil)
}

func (c *Client) CreateVolume(ctx context.Context, config DockerVolumeCreateJSONRequestBody) (Volume, error) {
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(config)
	if err != nil {
		return Volume{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "/v1.41/volumes/create", body)
	if err != nil {
		return Volume{}, err
	}

	out := Volume{}
	err = c.Do(req, &out)
	if err != nil {
		return Volume{}, err
	}

	return out, nil
}
