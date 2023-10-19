package api

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/mattn/go-tty"
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

type SSHParams struct {
	PodID   string
	Env     []string
	User    string
	WorkDir string
}

func (c *Client) SSH(ctx context.Context, tty *tty.TTY, params SSHParams) error {
	buf := &bytes.Buffer{}
	execID, err := c.initSSH(ctx, params, buf)
	if err != nil {
		return err
	}
	buf.Reset()

	req, err := c.newSSHRequest(ctx, execID, buf)
	if err != nil {
		return err
	}

	host := c.baseURL.Host
	if c.baseURL.Port() == "" {
		host = host + ":443"
	}

	var conn net.Conn

	if c.baseURL.Scheme == "https" {
		d := tls.Dialer{}
		conn, err = d.DialContext(ctx, "tcp", host)
		if err != nil {
			return err
		}
		conn.(*tls.Conn).NetConn().(*net.TCPConn).SetKeepAlive(true)
		conn.(*tls.Conn).NetConn().(*net.TCPConn).SetKeepAlivePeriod(time.Second * 2)
		defer conn.Close()
	} else {
		conn, err = net.Dial("tcp", host)
		if err != nil {
			return err
		}
		conn.(*net.TCPConn).SetKeepAlive(true)
		conn.(*net.TCPConn).SetKeepAlivePeriod(time.Second * 2)
		defer conn.Close()
	}

	err = req.Write(conn)
	if err != nil {
		return err
	}

	r, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return err
	}

	if r.StatusCode != 101 {
		return fmt.Errorf("unexpected status code %d", r.StatusCode)
	}

	// we must resize after acquiring the exec stream
	err = c.resizeSSH(ctx, tty, execID)
	if err != nil {
		log.Println("error resizing tty", err)
	}

	// set calling terminal raw mode
	restore, err := tty.Raw()
	if err == nil {
		defer restore()
	}

	go io.Copy(conn, tty.Input())
	_, err = io.Copy(tty.Output(), conn)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) resizeSSH(ctx context.Context, tty *tty.TTY, execID string) error {
	x, y, err := tty.Size()
	if err != nil {
		return err
	}

	resizeReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		path.Join("/v1.41/exec", execID, "resize?"+url.Values{
			"h": []string{strconv.Itoa(y)},
			"w": []string{strconv.Itoa(x)},
		}.Encode(),
		),
		nil,
	)
	if err != nil {
		return err
	}
	err = c.Do(resizeReq, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) initSSH(ctx context.Context, params SSHParams, buf *bytes.Buffer) (string, error) {
	err := json.NewEncoder(buf).Encode(types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		User:         params.User,
		Env:          params.Env,
		WorkingDir:   params.WorkDir,
		Cmd:          []string{"/bin/sh", "-c", "bash || sh"},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		path.Join("/v1.41/containers", params.PodID, "exec"),
		buf,
	)
	if err != nil {
		return "", err
	}

	response := types.IDResponse{}
	err = c.Do(req, &response)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func (c *Client) newSSHRequest(ctx context.Context, execID string, buf *bytes.Buffer) (*http.Request, error) {
	err := json.NewEncoder(buf).Encode(types.ExecStartCheck{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		path.Join("/v1.41/exec", execID, "start"),
		buf,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Upgrade", "tcp")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Host = c.baseURL.Host

	return req, nil
}
