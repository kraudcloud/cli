package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) CreateIDP(ctx context.Context,
	name string,
	namespace string,
	protocol string,
	metadata string,

) (*KraudIdentityProvider, error) {

	body := &KraudIdentityProvider{
		Name:        name,
		Namespace:   namespace,
		Protocol:    protocol,
		SvcMetadata: &metadata,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"/kr1/idp",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	var response = &KraudIdentityProvider{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) ListIDPs(ctx context.Context) (*KraudIdentityProviderList, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/kr1/idp",
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudIdentityProviderList{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) InspectIDP(ctx context.Context, id string) (*KraudIdentityProvider, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/kr1/idp/"+id,
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudIdentityProvider{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) DeleteIDP(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		"/kr1/idp/"+id,
		nil,
	)

	if err != nil {
		return err
	}

	return c.Do(req, nil)
}
