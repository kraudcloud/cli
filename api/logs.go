package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
)

// LogsOptions represents the options for the GetLogs method.
//
// currently, all but Follow are unimplemented.
type LogsOptions struct {
	Follow     bool `query:"follow"`
	Stdout     bool `query:"stdout"`
	Stderr     bool `query:"stderr"`
	Since      int  `query:"since"`
	Until      int  `query:"until"`
	Timestamps bool `query:"timestamps"`
	TailLines  int  `query:"tail"`
}

func (c *Client) GetLogs(ctx context.Context, container string, options LogsOptions) (io.ReadCloser, error) {
	u := url.URL{
		Path:     "/v1.41/containers/" + container + "/logs",
		RawQuery: structToQuery(options).Encode(),
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		u.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.DoRaw(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

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
