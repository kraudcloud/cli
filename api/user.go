package api

import (
	"context"
	"net/http"
)

func (c *Client) GetUser(ctx context.Context, name string) (*K8sUser, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/users/"+name,
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &K8sUser{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) GetUserMe(ctx context.Context) (*K8sUser, error) {
	return c.GetUser(ctx, ".me")
}
