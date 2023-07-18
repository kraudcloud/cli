package api

import (
	"context"
	"net/http"
	"path"
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
