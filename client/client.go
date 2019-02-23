package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client/runner"
	"github.com/ubclaunchpad/inertia/common"
)

// Client manages a deployment
type Client struct {
	version string
	out     io.Writer

	Remote *cfg.Remote

	ssh runner.SSHSession
}

// Options denotes configuration options for a Client
type Options struct {
	SSH runner.SSHOptions
	Out io.Writer
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

// GetSSHClient instantiates an SSH client for Inertia-related commands
func (c *Client) GetSSHClient() (*SSHClient, error) {
	if c.ssh == nil {
		return nil, errors.New("client not configured for SSH access")
	}
	return &SSHClient{
		ssh:    c.ssh,
		remote: c.Remote,
	}, nil
}

// Up brings the project up on the remote VPS instance specified
// in the deployment object.
func (c *Client) Up(project, url string, profile cfg.Profile, stream bool) (*http.Response, error) {
	return c.post("/up", &api.UpRequest{
		Stream:        stream,
		Project:       project,
		WebHookSecret: c.Remote.Daemon.WebHookSecret,

		BuildType:     string(profile.Build.Type),
		BuildFilePath: profile.Build.BuildFilePath,
		GitOptions: api.GitOptions{
			RemoteURL: common.GetSSHRemoteURL(url),
			Branch:    profile.Branch,
		},
	})
}

// LogIn gets an access token for the user with the given credentials. Use ""
// for totp if none is required.
func (c *Client) LogIn(user, password, totp string) (*http.Response, error) {
	return c.post("/user/login", &api.UserRequest{
		Username: user,
		Password: password,
		Totp:     totp,
	})
}

// Token generates token on this remote.
func (c *Client) Token() (*http.Response, error) {
	return c.get("/token", nil)
}

// Prune clears Docker ReadFiles on this remote.
func (c *Client) Prune() (*http.Response, error) {
	return c.post("/prune", nil)
}

// Down brings the project down on the remote VPS instance specified
// in the configuration object.
func (c *Client) Down() (*http.Response, error) {
	return c.post("/down", nil)
}

// Status lists the currently active containers on the remote VPS instance
func (c *Client) Status() (*http.Response, error) {
	resp, err := c.get("/status", nil)
	if err != nil &&
		(strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "refused")) {
		return nil, errors.New("daemon on remote appears offline or inaccessible")
	}
	return resp, err
}

// Reset shuts down deployment and deletes the contents of the deployment's
// project directory
func (c *Client) Reset() (*http.Response, error) {
	return c.post("/reset", nil)
}

// Logs get logs of given container
func (c *Client) Logs(container string, entries int) (*http.Response, error) {
	reqContent := map[string]string{api.Container: container}
	if entries > 0 {
		reqContent[api.Entries] = strconv.Itoa(entries)
	}

	return c.get("/logs", reqContent)
}

// LogsWebSocket opens a websocket connection to given container's logs
func (c *Client) LogsWebSocket(container string, entries int) (SocketReader, error) {
	addr, err := c.Remote.DaemonAddr()
	if err != nil {
		return nil, err
	}
	host, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	// Set up request
	url := &url.URL{Scheme: "wss", Host: host.Host, Path: "/logs"}
	params := map[string]string{
		api.Container: container,
		api.Stream:    "true",
	}
	if entries > 0 {
		params[api.Entries] = strconv.Itoa(entries)
	}
	encodeQuery(url, params)

	// Set up authorization
	var header = http.Header{}
	header.Set("Authorization", "Bearer "+c.Remote.Daemon.Token)

	// Attempt websocket connection
	socket, resp, err := buildWebSocketDialer(c.Remote.Daemon.VerifySSL).
		Dial(url.String(), header)
	if err == websocket.ErrBadHandshake {
		return nil, fmt.Errorf("websocket handshake failed with status %d", resp.StatusCode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon at %s: %s", url.Host, err.Error())
	}
	return socket, nil
}

// UpdateEnv updates environment variable
func (c *Client) UpdateEnv(name, value string, encrypt, remove bool) (*http.Response, error) {
	return c.post("/env", api.EnvRequest{
		Name: name, Value: value, Encrypt: encrypt, Remove: remove,
	})
}

// ListEnv lists environment variables currently set on remote
func (c *Client) ListEnv() (*http.Response, error) {
	return c.get("/env", nil)
}

// AddUser adds an authorized user for access to Inertia Web
func (c *Client) AddUser(username, password string, admin bool) (*http.Response, error) {
	return c.post("/user/add", &api.UserRequest{
		Username: username,
		Password: password,
		Admin:    admin,
	})
}

// RemoveUser prevents a user from accessing Inertia Web
func (c *Client) RemoveUser(username string) (*http.Response, error) {
	return c.post("/user/remove", &api.UserRequest{Username: username})
}

// ResetUsers resets all users on the remote.
func (c *Client) ResetUsers() (*http.Response, error) {
	return c.post("/user/reset", nil)
}

// ListUsers lists all users on the remote.
func (c *Client) ListUsers() (*http.Response, error) {
	return c.get("/user/list", nil)
}

// EnableTotp enables Totp for a given user
func (c *Client) EnableTotp(username, password string) (*http.Response, error) {
	return c.post("/user/totp/enable", &api.UserRequest{
		Username: username,
		Password: password,
	})
}

// DisableTotp disables Totp for a given user
func (c *Client) DisableTotp() (*http.Response, error) {
	return c.post("/user/totp/disable", nil)
}

// Sends a GET request. "queries" contains query string arguments.
func (c *Client) get(endpoint string, queries map[string]string) (*http.Response, error) {
	// Assemble request
	req, err := c.buildRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add query strings
	if queries != nil {
		encodeQuery(req.URL, queries)
	}

	return buildHTTPSClient(c.Remote.Daemon.VerifySSL).Do(req)
}

func (c *Client) post(endpoint string, requestBody interface{}) (*http.Response, error) {
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
	req, err := c.buildRequest("POST", endpoint, payload)
	if err != nil {
		return nil, err
	}

	return buildHTTPSClient(c.Remote.Daemon.VerifySSL).Do(req)
}

func (c *Client) buildRequest(method string, endpoint string, payload io.Reader) (*http.Request, error) {
	// Assemble URL
	addr, err := c.Remote.DaemonAddr()
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, endpoint)

	// Assemble request
	req, err := http.NewRequest(method, url.String(), payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Remote.Daemon.Token)

	return req, nil
}
