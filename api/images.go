package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) ListImages(ctx context.Context) (*KraudImageNameList, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/images",
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudImageNameList{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) InspectImage(ctx context.Context, q string) (*KraudImageName, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/images/"+q,
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudImageName{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) CreateImage(ctx context.Context, body CreateImageJSONBody) (*KraudCreateImageResponse, error) {

	jq, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"/apis/kraudcloud.com/v1/images",
		bytes.NewBuffer(jq),
	)

	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	var response = &KraudCreateImageResponse{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
