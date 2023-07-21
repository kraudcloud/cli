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

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		path.Join("/apis/kraudcloud.com/v1/feeds", feedID, "apps", appID, "launch"),
		buf,
	)
	if err != nil {
		return nil, err
	}

	var response KraudLaunchAppResponse
	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
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
		t, msgReader, err := conn.Reader(ctx)
		if err != nil {
			return err
		}

		if t != websocket.MessageText {
			continue
		}

		err = CopyWS(w, msgReader)
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

func CopyWS(w io.Writer, r io.Reader) error {
	decoder := json.NewDecoder(r)

	for {
		var msg json.RawMessage
		err := decoder.Decode(&msg)
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		}

		// check for log
		var log KraudWSLaunchLog
		err = json.Unmarshal(msg, &log)
		if err == nil && log.Log != "" {
			_, err = fmt.Fprintln(w, strings.TrimSpace(log.Log))
			if err != nil {
				return err
			}
			continue
		}

		var meta KraudWSLaunchMeta
		err = json.Unmarshal(msg, &meta)
		if err != nil {
			return err
		}
	}
}
