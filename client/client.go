package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client/runner"
	"github.com/ubclaunchpad/inertia/common"
)

// Client manages a deployment
type Client struct {
	om  sync.Mutex
	out io.Writer

	ssh   runner.SSHSession
	debug bool

	Remote *cfg.Remote
}

// Options denotes configuration options for a Client
type Options struct {
	SSH   runner.SSHOptions
	Out   io.Writer
	Debug bool
}

// NewClient sets up a client to communicate to the daemon at the given remote
func NewClient(remote *cfg.Remote, opts Options) *Client {
	if opts.Out == nil {
		opts.Out = common.DevNull{}
	}

	if remote.Version == "" {
		remote.Version = "latest"
	}

	return &Client{
		out:    opts.Out,
		Remote: remote,
		ssh:    runner.NewSSHRunner(remote.IP, remote.SSH, opts.SSH),
	}
}

// WithWriter sets the given io.Writer as the client's default output
func (c *Client) WithWriter(out io.Writer) { c.out = out }

// WithDebug sets the client's debug mode
func (c *Client) WithDebug(debug bool) { c.debug = debug }

// GetSSHClient instantiates an SSH client for Inertia-related commands
func (c *Client) GetSSHClient() (*SSHClient, error) {
	if c.ssh == nil {
		return nil, errors.New("client not configured for SSH access")
	}
	return &SSHClient{
		ssh:    c.ssh,
		remote: c.Remote,
		debug:  c.debug,
		out:    c.out,
	}, nil
}

// GetUserClient instantiates an API client for Inertia user management commands
func (c *Client) GetUserClient() *UserClient {
	return NewUserClient(c)
}

// UpRequest declares parameters for project deployment
type UpRequest struct {
	Project string
	URL     string
	Profile cfg.Profile
}

// Up brings the project up on the remote VPS instance specified
// in the deployment object.
func (c *Client) Up(ctx context.Context, req UpRequest) error {
	notif := req.Profile.Notifiers
	if notif == nil {
		notif = &cfg.Notifiers{}
	}

	resp, err := c.post(ctx, "/up", &api.UpRequest{
		Stream:        false,
		Project:       req.Project,
		WebHookSecret: c.Remote.Daemon.WebHookSecret,
		BuildType:     string(req.Profile.Build.Type),
		BuildFilePath: req.Profile.Build.BuildFilePath,
		GitOptions: api.GitOptions{
			RemoteURL: common.GetSSHRemoteURL(req.URL),
			Branch:    req.Profile.Branch,
		},
		IntermediaryContainers: req.Profile.Build.IntermediaryContainers,
		SlackNotificationURL:   notif.SlackNotificationURL,
	})
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err.Error())
	}
	base, err := c.unmarshal(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err.Error())
	}
	return base.Error()
}

// UpWithOutput blocks and streams 'up' output to the client's io.Writer
func (c *Client) UpWithOutput(ctx context.Context, req UpRequest) error {
	resp, err := c.post(ctx, "/up", &api.UpRequest{
		Stream:        true,
		Project:       req.Project,
		WebHookSecret: c.Remote.Daemon.WebHookSecret,
		BuildType:     string(req.Profile.Build.Type),
		BuildFilePath: req.Profile.Build.BuildFilePath,
		GitOptions: api.GitOptions{
			RemoteURL: common.GetSSHRemoteURL(req.URL),
			Branch:    req.Profile.Branch,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err.Error())
	}
	defer resp.Body.Close()

	// read until error
	var scan = bufio.NewScanner(resp.Body)
	var errC = make(chan error, 1)
	go func() {
		for scan.Scan() {
			c.om.Lock()
			fmt.Fprintln(c.out, scan.Text())
			c.om.Unlock()
		}
		if err := scan.Err(); err != nil {
			errC <- fmt.Errorf("error occured while reading output: %s", err.Error())
			return
		}
	}()

	// block until done
	for {
		select {
		case <-ctx.Done():
			c.debugf("context cancelled, closing connection")
			return nil
		case err := <-errC:
			c.debugf("error received: %s", err.Error())
			return err
		}
	}
}

// Token generates token on this remote.
func (c *Client) Token(ctx context.Context) (token string, err error) {
	resp, err := c.get(ctx, "/token", nil)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %s", err.Error())
	}

	base, err := c.unmarshal(resp.Body, api.KV{Key: "token", Value: &token})
	resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("failed to read response: %s", err.Error())
	}

	return token, base.Error()
}

// Prune clears Docker ReadFiles on this remote.
func (c *Client) Prune(ctx context.Context) error {
	resp, err := c.post(ctx, "/prune", nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err.Error())
	}

	base, err := c.unmarshal(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err.Error())
	}

	return base.Error()
}

// Down brings the project down on the remote VPS instance specified
// in the configuration object.
func (c *Client) Down(ctx context.Context) error {
	resp, err := c.post(ctx, "/down", nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err.Error())
	}

	base, err := c.unmarshal(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err.Error())
	}

	return base.Error()
}

// Status lists the currently active containers on the remote VPS instance
func (c *Client) Status(ctx context.Context) (*api.DeploymentStatus, error) {
	resp, err := c.get(ctx, "/status", nil)
	if err != nil &&
		(strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "refused")) {
		return nil, errors.New("daemon on remote appears offline or inaccessible")
	} else if err != nil {
		return nil, fmt.Errorf("failed to make request: %s", err.Error())
	}

	var status = &api.DeploymentStatus{}
	base, err := c.unmarshal(resp.Body, api.KV{
		Key: "status", Value: status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %s", err.Error())
	}

	return status, base.Error()
}

// Reset shuts down deployment and deletes the contents of the deployment's
// project directory
func (c *Client) Reset(ctx context.Context) error {
	resp, err := c.post(ctx, "/reset", nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err.Error())
	}

	base, err := c.unmarshal(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err.Error())
	}

	return base.Error()
}

// LogsRequest denotes parameters for log querying
type LogsRequest struct {
	Container string
	Entries   int
}

// Logs get logs of given container
func (c *Client) Logs(ctx context.Context, req LogsRequest) ([]string, error) {
	reqContent := map[string]string{api.Container: req.Container}
	if req.Entries > 0 {
		reqContent[api.Entries] = strconv.Itoa(req.Entries)
	}

	resp, err := c.get(ctx, "/logs", reqContent)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %s", err.Error())
	}

	var logs = make([]string, 0)
	b, err := c.unmarshal(resp.Body, api.KV{Key: "logs", Value: &logs})
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %s", err.Error())
	}

	return logs, b.Error()
}

// LogsWithOutput opens a websocket connection to given container's logs and
// streams it to the given io.Writer
func (c *Client) LogsWithOutput(ctx context.Context, req LogsRequest) error {
	addr, err := c.Remote.DaemonAddr()
	if err != nil {
		return err
	}
	host, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("invalid daemon address: %s", err.Error())
	}

	// Set up request
	var url = &url.URL{Scheme: "wss", Host: host.Host, Path: "/logs"}
	var params = map[string]string{
		api.Container: req.Container,
		api.Stream:    "true",
	}
	if req.Entries > 0 {
		params[api.Entries] = strconv.Itoa(req.Entries)
	}
	encodeQuery(url, params)

	// Set up authorization
	var header = http.Header{}
	header.Set("Authorization", "Bearer "+c.Remote.Daemon.Token)

	// set up websocket connection
	c.debugf("request constructed: %s (authorized: %v, verified: %v)",
		url.String(), c.Remote.Daemon.Token != "", c.Remote.Daemon.VerifySSL)
	socket, resp, err := buildWebSocketDialer(c.Remote.Daemon.VerifySSL).
		DialContext(ctx, url.String(), header)
	if err == websocket.ErrBadHandshake {
		return fmt.Errorf("websocket handshake failed with status %d", resp.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %s", err.Error())
	}
	defer socket.Close()
	c.debugf("websocket connection established")

	// read from socket until error
	var errC = make(chan error, 1)
	go func() {
		for {
			_, line, err := socket.ReadMessage()
			if err != nil {
				errC <- fmt.Errorf("error occured while reading from socket: %s", err.Error())
				return
			}
			c.om.Lock()
			fmt.Fprint(c.out, string(line))
			c.om.Unlock()
		}
	}()

	// block until done
	for {
		select {
		case <-ctx.Done():
			c.debugf("context cancelled, closing connection")
			return nil
		case err := <-errC:
			c.debugf("error received: %s", err.Error())
			return err
		}
	}
}

// UpdateEnv updates environment variable
func (c *Client) UpdateEnv(ctx context.Context, name, value string, encrypt, remove bool) error {
	resp, err := c.post(ctx, "/env", api.EnvRequest{
		Name: name, Value: value, Encrypt: encrypt, Remove: remove,
	})
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err.Error())
	}

	base, err := c.unmarshal(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err.Error())
	}

	return base.Error()
}

// ListEnv lists environment variables currently set on remote
func (c *Client) ListEnv(ctx context.Context) ([]string, error) {
	resp, err := c.get(ctx, "/env", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %s", err.Error())
	}

	var variables = make([]string, 0)
	base, err := c.unmarshal(resp.Body, api.KV{Key: "variables", Value: &variables})
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %s", err.Error())
	}

	return variables, base.Error()
}

// Sends a GET request. "queries" contains query string arguments.
func (c *Client) get(
	ctx context.Context,
	endpoint string,
	queries map[string]string,
) (*http.Response, error) {
	// Assemble request
	req, err := c.buildRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add query strings
	if queries != nil {
		encodeQuery(req.URL, queries)
	}

	return buildHTTPSClient(c.Remote.Daemon.VerifySSL).Do(req)
}

func (c *Client) post(
	ctx context.Context,
	endpoint string,
	requestBody interface{},
) (*http.Response, error) {
	// Assemble payload
	var payload io.Reader
	if requestBody != nil {
		body, err := json.Marshal(requestBody)
		if err != nil {
			return nil, err
		}
		payload = bytes.NewReader(body)
	} else {
		payload = nil
	}

	// Assemble request
	req, err := c.buildRequest(ctx, "POST", endpoint, payload)
	if err != nil {
		return nil, err
	}

	return buildHTTPSClient(c.Remote.Daemon.VerifySSL).Do(req)
}

func (c *Client) buildRequest(
	ctx context.Context,
	method string,
	endpoint string,
	payload io.Reader,
) (*http.Request, error) {
	// Assemble URL
	addr, err := c.Remote.DaemonAddr()
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid url configuration: %s", err.Error())
	}
	url.Path = path.Join(url.Path, endpoint)

	// Assemble request
	req, err := http.NewRequest(method, url.String(), payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Remote.Daemon.Token)
	req.WithContext(ctx)
	c.debugf("request constructed: %s %s (authorized: %v, with payload: %v, verified: %v)",
		method, req.URL.String(), c.Remote.Daemon.Token != "", payload != nil, c.Remote.Daemon.VerifySSL)

	return req, nil
}

// unmarshal wraps c.unmarshal and logs responses
func (c *Client) unmarshal(r io.Reader, kvs ...api.KV) (*api.BaseResponse, error) {
	b, err := api.Unmarshal(r, kvs...)
	if b != nil {
		bytes, _ := json.MarshalIndent(b, "", "  ")
		c.debugf("response received: %s (%s)", b.Message, string(bytes))
	}
	return b, err
}

// debugf logs to the client's output if debug is enabled
func (c *Client) debugf(format string, args ...interface{}) {
	if c.debug {
		c.om.Lock()
		fmt.Fprintf(c.out, "DEBUG: "+format+"\n", args...)
		c.om.Unlock()
	}
}
