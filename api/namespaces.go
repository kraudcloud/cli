package api

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

func (c *Client) ListNamespaces(ctx context.Context) (*K8sNamespaceList, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/api/v1/namespaces",
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &K8sNamespaceList{}
	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) DeleteNamespace(ctx context.Context, namespace string, force bool) error {
	u := url.URL{
		Path: path.Join("/api/v1/namespaces/", namespace),
		RawQuery: url.Values{
			"recursive": []string{strconv.FormatBool(force)},
		}.Encode(),
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		u.String(),
		nil,
	)

	if err != nil {
		return err
	}

	err = c.Do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) NamespaceOverview(ctx context.Context, namespace string) (KrNamespaceOverview, error) {
	path, err := url.JoinPath("/apis/kr/v1/namespaces/by-name", namespace, "overview.json")
	if err != nil {
		return KrNamespaceOverview{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		path,
		nil,
	)

	if err != nil {
		return KrNamespaceOverview{}, err
	}

	var response KrNamespaceOverview
	err = c.Do(req, &response)
	if err != nil {
		return KrNamespaceOverview{}, err
	}

	return response, nil
}
