package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	dockerclient "github.com/docker/docker/client"
)

var Version string = "devel"

type Client struct {
	// HTTP client used to communicate with the API.
	HTTPClient *http.Client

	// AuthToken is the token used to authenticate with the API.
	AuthToken string

	BaseURL *url.URL
}

func NewClient(authToken string, baseURL *url.URL) *Client {
	userAgent := fmt.Sprintf("kra %s", Version)

	client := &http.Client{
		Transport: &kraTransport{
			Base:      http.DefaultTransport,
			BaseURL:   baseURL,
			UserAgent: userAgent,
			AuthToken: authToken,
		},
	}

	return &Client{
		HTTPClient: client,
		AuthToken:  authToken,
		BaseURL:    baseURL,
	}
}

func (c *Client) Close() {
	c.HTTPClient.CloseIdleConnections()
}

func (c *Client) DockerClient() *dockerclient.Client {
	dc, err := dockerclient.NewClientWithOpts(
		dockerclient.WithAPIVersionNegotiation(),
		dockerclient.WithHost("tcp://"+c.BaseURL.Host),
		dockerclient.WithHTTPClient(c.HTTPClient),
	)
	if err != nil {
		panic(err)
	}

	return dc
}

func (c *Client) Do(req *http.Request, response interface{}) error {
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTPClient.Do(req)

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
	return c.HTTPClient.Do(req)
}

type kraTransport struct {
	Base http.RoundTripper

	BaseURL *url.URL

	UserAgent string

	AuthToken string
}

func (t *kraTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.BaseURL.Scheme
	req.URL.Host = t.BaseURL.Host
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", t.UserAgent)
	}

	if t.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.AuthToken))
	}

	return t.Base.RoundTrip(req)
}
