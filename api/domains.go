package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) CreateDomain(ctx context.Context, d KraudDomainCreate) (*KraudDomain, error) {
	b := &bytes.Buffer{}
	err := json.NewEncoder(b).Encode(d)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"/apis/kraudcloud.com/v1/domains",
		b,
	)
	if err != nil {
		return nil, err
	}

	out := &KraudDomain{}
	err = c.Do(req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
