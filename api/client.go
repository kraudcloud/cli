package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	dockerclient "github.com/docker/docker/client"
)

var Version string = "devel"

type Client struct {
	// HTTP client used to communicate with the API.
	HTTPClient *http.Client

	// AuthToken is the token used to authenticate with the API.
	AuthToken string

	Host string
}

func NewClient(authToken string) *Client {
	host := "api.kraudcloud.com"
	userAgent := fmt.Sprintf("kra %s", Version)

	client := &http.Client{
		Transport: &kraTransport{
			Base:      http.DefaultTransport,
			Scheme:    "https",
			Host:      host,
			UserAgent: userAgent,
			AuthToken: authToken,
		},
	}

	return &Client{
		HTTPClient: client,
		AuthToken:  authToken,
		Host:       host,
	}
}

func (c *Client) Close() {
	c.HTTPClient.CloseIdleConnections()
}

func (c *Client) DockerClient() *dockerclient.Client {
	dc, err := dockerclient.NewClientWithOpts(
		dockerclient.WithAPIVersionNegotiation(),
		dockerclient.WithHost("tcp://"+c.Host),
		dockerclient.WithHTTPClient(c.HTTPClient),
	)
	if err != nil {
		panic(err)
	}

	return dc
}

func (c *Client) Do(req *http.Request, response interface{}) error {

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

	Scheme string
	Host   string

	UserAgent string

	AuthToken string
}

func (t *kraTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.Scheme
	req.URL.Host = t.Host
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", t.UserAgent)
	}

	if t.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.AuthToken))
	}

	return t.Base.RoundTrip(req)
}
