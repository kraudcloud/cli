package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"

	"nhooyr.io/websocket"
)

func (c *Client) ListFeeds(ctx context.Context) (KraudFeedList, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/feeds",
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response KraudFeedList
	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

type ListAppsResponse struct {
	Items []KraudAppOverview `json:"apps"`
}

func (c *Client) ListApps(ctx context.Context, feedID string) (*ListAppsResponse, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		path.Join("/apis/kraudcloud.com/v1/feeds", feedID, "apps"),
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response ListAppsResponse
	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) Launch(ctx context.Context, template string, config KraudLaunchSettings) (*KraudLaunchAppResponse, error) {
	buf := bytes.Buffer{}

	mw := multipart.NewWriter(&buf)
	err := mw.WriteField("config", MustJSONString(config))
	if err != nil {
		return nil, err
	}

	err = mw.WriteField("template", template)
	if err != nil {
		return nil, err
	}

	err = mw.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"/apis/kraudcloud.com/v1/launch",
		&buf,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())

	var response KraudLaunchAppResponse
	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) LaunchApp(ctx context.Context, feedID string, appID string, params KraudLaunchSettings) (*KraudLaunchAppResponse, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(params)
	if err != nil {
		return nil, err
	}

	path, err := url.JoinPath("/apis/kraudcloud.com/v1/feeds", feedID, "apps", appID, "launch")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		path,
		buf,
	)
	if err != nil {
		return nil, err
	}

	resp := &KraudLaunchAppResponse{}
	err = c.Do(req, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) LaunchAttach(ctx context.Context, w io.Writer, launchID string) error {
	u := &url.URL{
		Scheme: c.BaseURL.Scheme,
		Host:   c.BaseURL.Host,
		Path:   path.Join("/apis/kraudcloud.com/v1/launch", launchID, "attach"),
	}

	conn, _, err := websocket.Dial(ctx, u.String(), &websocket.DialOptions{
		HTTPClient: c.HTTPClient,
	})
	if err != nil {
		return err
	}

	for {
		t, msg, err := conn.Read(ctx)
		if err != nil {
			return err
		}

		if t != websocket.MessageText {
			continue
		}

		err = wsMessage(w, msg)
		if err != nil {
			return err
		}
	}
}

func MustJSONString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	return string(b)
}

func wsMessage(w io.Writer, msg []byte) error {

	// check for log
	var log KraudWSLaunchLog
	err := json.Unmarshal(msg, &log)
	if err == nil && log.Log != "" {
		_, err = fmt.Fprintln(w, strings.TrimSpace(log.Log))
		if err != nil {
			return err
		}

		return nil
	}

	var meta KraudWSLaunchMeta
	err = json.Unmarshal(msg, &meta)
	if err != nil {
		return err
	}

	switch {
	case meta.Error != nil:
		fmt.Fprintf(w, "Error: %s\n", *meta.Error)
	case meta.DeploymentAids != nil:
		fmt.Fprintf(w, "Deployment AIDs: %v\n", meta.DeploymentAids)
	}

	return nil
}
