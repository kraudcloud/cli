package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {

	// HTTP client used to communicate with the API.
	HttpClient *http.Client

	// AuthToken is the token used to authenticate with the API.
	AuthToken string
}

func NewClient(authToken string) *Client {
	return &Client{
		HttpClient: http.DefaultClient,
		AuthToken:  authToken,
	}
}

func (c *Client) Close() {
	c.HttpClient.CloseIdleConnections()
}

func (c *Client) Do(req *http.Request, response interface{}) error {

	req.URL.Host = "api.kraudcloud.com"
	req.URL.Scheme = "https"

	req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	resp, err := c.HttpClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 201 {
		var ee ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&ee)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s", ee.Message)
	}

	if response != nil {
		err := json.NewDecoder(resp.Body).Decode(response)
		if err != nil {
			return err
		}
	}

	return nil
}
