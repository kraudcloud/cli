package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
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

	// req.URL.Host = "localhost:3804"
	// req.URL.Scheme = "http"

	req.Header.Set("Authorization", "Bearer "+c.AuthToken)

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "kra v1.0.1")
	resp, err := c.HttpClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		var ee ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&ee)
		if err != nil {
			return fmt.Errorf("%s", resp.Status)
		}
		if ee.Message != "" {
			return fmt.Errorf("%s", ee.Message)
		}
		return fmt.Errorf("%s", resp.Status)
	}

	if response != nil {
		err := json.NewDecoder(resp.Body).Decode(response)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DoRaw(req *http.Request) (*http.Response, error) {

	req.URL.Host = "api.kraudcloud.com"
	req.URL.Scheme = "https"

	req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	return c.HttpClient.Do(req)
}

// structToQuery converts a struct to a url.Values
//
// it uses the "query" tag on each field to determine the key.
func structToQuery(v interface{}) url.Values {
	values := url.Values{}

	rv := reflect.ValueOf(v)
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if field.IsZero() {
			continue
		}

		values.Set(rv.Type().Field(i).Tag.Get("query"), fmt.Sprintf("%v", field))
	}

	return values
}
