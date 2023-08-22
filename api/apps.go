package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/kraudcloud/cli/compose/envparser"
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

func (c *Client) CreateFeed(ctx context.Context, kcf KraudCreateFeed) (*KraudFeed, error) {
	buf := &bytes.Buffer{}
	json.NewEncoder(buf).Encode(kcf)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"/apis/kraudcloud.com/v1/feeds",
		buf,
	)
	if err != nil {
		return nil, err
	}

	var response KraudFeed
	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) PushApp(ctx context.Context, feedID string, template io.Reader, changelog string) error {
	buf := &bytes.Buffer{}

	body := multipart.NewWriter(buf)
	file, err := body.CreateFormFile("template", "template.yaml")
	if err != nil {
		return fmt.Errorf("error creating form file: %w", err)
	}

	io.Copy(file, template)
	if changelog != "" {
		err = body.WriteField("changelog", changelog)
		if err != nil {
			return err
		}
	}

	err = body.Close()
	if err != nil {
		return fmt.Errorf("error closing multipart writer: %w", err)
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("/apis/kraudcloud.com/v1/feeds/%s/app", feedID),
		buf,
	)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", body.FormDataContentType())

	err = c.Do(req, nil)
	if err != nil {
		return fmt.Errorf("error pushing app: %w", err)
	}

	return nil
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

func (c *Client) InspectApp(ctx context.Context, feedID string, appID string) (*KraudAppOverview, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		path.Join("/apis/kraudcloud.com/v1/feeds", feedID, "apps", appID, "template"),
		nil,
	)
	if err != nil {
		return nil, err
	}

	response := KraudAppOverview{}
	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type LaunchParams struct {
	Template  string
	Env       map[string]string
	Namespace string
	Detach    bool
}

func (c *Client) Launch(ctx context.Context, lp LaunchParams, response io.Writer) error {
	u := &url.URL{
		Path: "/apis/kraudcloud.com/v1/launch",
		RawQuery: url.Values{
			"namespace": []string{lp.Namespace},
		}.Encode(),
	}

	buf := bytes.Buffer{}

	mw := multipart.NewWriter(&buf)
	err := mw.WriteField("docker-compose.yml", lp.Template)
	if err != nil {
		return err
	}

	ew, err := mw.CreateFormField(".env")
	if err != nil {
		return err
	}
	err = envparser.EncodeEnv(ew, lp.Env)
	if err != nil {
		return err
	}

	err = mw.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		u.String(),
		&buf,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())

	if lp.Detach {
		req.Header.Set("Accept", "application/json")
	}

	resp, err := c.DoRaw(req)
	if err != nil {
		return err
	}

	_, err = io.Copy(response, resp.Body)
	if errors.Is(err, io.EOF) {
		return nil
	}

	return err
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
		Scheme: c.baseURL.Scheme,
		Host:   c.baseURL.Host,
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
		if errors.Is(err, io.EOF) {
			return nil
		}

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
