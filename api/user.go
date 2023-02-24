package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) GetUser(ctx context.Context, uuid string) (*K8sUser, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/users/"+uuid,
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

func (c *Client) RotateUserCredentials(ctx context.Context, uuid string, name string, context string) (io.ReadCloser, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"/apis/kraudcloud.com/v1/users/"+uuid+"/credentials/"+name+"/rotate?context="+context,
		nil,
	)

	if err != nil {
		return nil, err
	}

	resp, err := c.DoRaw(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		var ee ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&ee)
		if err != nil {
			return nil, fmt.Errorf("%s", resp.Status)
		}
		if ee.Message != "" {
			return nil, fmt.Errorf("%s", ee.Message)
		}
		return nil, fmt.Errorf("%s", resp.Status)
	}

	return resp.Body, nil
}
