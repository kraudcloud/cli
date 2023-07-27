package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) ListPods(ctx context.Context, withStatus bool) (*KraudPodList, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/pods?status="+fmt.Sprint(withStatus),
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudPodList{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) InspectPod(ctx context.Context, search string) (*KraudPod, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"/apis/kraudcloud.com/v1/pods/"+search,
		nil,
	)

	if err != nil {
		return nil, err
	}

	var response = &KraudPod{}
	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) EditPod(ctx context.Context, search string, pod *KraudPod) error {

	body, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"PUT",
		"/apis/kraudcloud.com/v1/pods/"+search,
		bytes.NewReader(body),
	)

	if err != nil {
		return err
	}

	var response interface{}
	err = c.Do(req, response)
	if err != nil {
		return err
	}

	return nil
}
