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

	// roundtrip options
	baseURL   *url.URL
	authToken string
	userAgent string
	base      http.RoundTripper
}

func NewClient(authToken string, baseURL *url.URL) *Client {
	userAgent := fmt.Sprintf("kra %s", Version)

	c := &Client{
		base:      http.DefaultTransport,
		baseURL:   baseURL,
		authToken: authToken,
		userAgent: userAgent,
	}

	client := &http.Client{
		Transport: c,
	}

	c.HTTPClient = client
	return c
}

func (c *Client) DockerClient() *dockerclient.Client {
	dc, err := dockerclient.NewClientWithOpts(
		dockerclient.WithAPIVersionNegotiation(),
		dockerclient.WithHost("tcp://"+c.baseURL.Host),
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

func (c *Client) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = c.baseURL.Scheme
	req.URL.Host = c.baseURL.Host
	if req.Header.Get("User-Agent") == "" && c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	if req.Header.Get("Authorization") == "" && c.authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	}

	return c.base.RoundTrip(req)
}
