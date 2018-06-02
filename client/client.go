package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/ubclaunchpad/inertia/common"
)

// Client manages a deployment
type Client struct {
	*RemoteVPS
	version   string
	project   string
	buildType string
	sshRunner SSHSession
}

// NewClient sets up a client to communicate to the daemon at
// the given named remote.
func NewClient(remoteName string, config *Config) (*Client, bool) {
	remote, found := config.GetRemote(remoteName)
	if !found {
		return nil, false
	}

	return &Client{
		RemoteVPS: remote,
		version:   config.Version,
		project:   config.Project,
		buildType: config.BuildType,
		sshRunner: NewSSHRunner(remote),
	}, true
}

// BootstrapRemote configures a remote vps for continuous deployment
// by installing docker, starting the daemon and building a
// public-private key-pair. It outputs configuration information
// for the user.
func (c *Client) BootstrapRemote(repoName string) error {
	println("Setting up remote \"" + c.Name + "\" at " + c.IP)

	println(">> Step 1/4: Installing docker...")
	err := c.installDocker(c.sshRunner)
	if err != nil {
		return err
	}

	println("\n>> Step 2/4: Building deploy key...")
	if err != nil {
		return err
	}
	pub, err := c.keyGen(c.sshRunner)
	if err != nil {
		return err
	}

	// This step needs to run before any other commands that rely on
	// the daemon image, since the daemon is loaded here.
	println("\n>> Step 3/4: Starting daemon...")
	if err != nil {
		return err
	}
	err = c.DaemonUp(c.version, c.IP, c.Daemon.Port)
	if err != nil {
		return err
	}

	println("\n>> Step 4/4: Fetching daemon API token...")
	token, err := c.getDaemonAPIToken(c.sshRunner, c.version)
	if err != nil {
		return err
	}
	c.Daemon.Token = token

	println("\nInertia has been set up and daemon is running on remote!")
	println("You may have to wait briefly for Inertia to set up some dependencies.")
	fmt.Printf("Use 'inertia %s logs --stream' to check on the daemon's setup progress.\n\n", c.Name)

	println("=============================\n")

	// Output deploy key to user.
	println(">> GitHub Deploy Key (add to https://www.github.com/" + repoName + "/settings/keys/new): ")
	println(pub.String())

	// Output Webhook url to user.
	println(">> GitHub WebHook URL (add to https://www.github.com/" + repoName + "/settings/hooks/new): ")
	println("WebHook Address:  https://" + c.IP + ":" + c.Daemon.Port + "/webhook")
	println("WebHook Secret:   " + c.Daemon.Secret)
	println(`Note that you will have to disable SSH verification in your webhook
settings - Inertia uses self-signed certificates that GitHub won't
be able to verify.` + "\n")

	println(`Inertia daemon successfully deployed! Add your webhook url and deploy
key to enable continuous deployment.`)
	fmt.Printf("Then run 'inertia %s up' to deploy your application.\n", c.Name)

	return nil
}

// DaemonUp brings the daemon up on the remote instance.
func (c *Client) DaemonUp(daemonVersion, host, daemonPort string) error {
	scriptBytes, err := Asset("client/bootstrap/daemon-up.sh")
	if err != nil {
		return err
	}

	// Run inertia daemon.
	daemonCmdStr := fmt.Sprintf(string(scriptBytes), daemonVersion, daemonPort, host)
	return c.sshRunner.RunStream(daemonCmdStr, false)
}

// DaemonDown brings the daemon down on the remote instance
func (c *Client) DaemonDown() error {
	scriptBytes, err := Asset("client/bootstrap/daemon-down.sh")
	if err != nil {
		return err
	}

	_, stderr, err := c.sshRunner.Run(string(scriptBytes))
	if err != nil {
		println(stderr.String())
		return err
	}

	return nil
}

// Up brings the project up on the remote VPS instance specified
// in the deployment object.
func (c *Client) Up(gitRemoteURL, buildType string, stream bool) (*http.Response, error) {
	if buildType == "" {
		buildType = c.buildType
	}

	reqContent := &common.DaemonRequest{
		Stream:    stream,
		Project:   c.project,
		BuildType: buildType,
		Secret:    c.RemoteVPS.Daemon.Secret,
		GitOptions: &common.GitOptions{
			RemoteURL: common.GetSSHRemoteURL(gitRemoteURL),
			Branch:    c.Branch,
		},
	}
	return c.post("/up", reqContent)
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
		return nil, fmt.Errorf("daemon on remote %s appears offline or inaccessible", c.Name)
	}
	return resp, err
}

// Reset shuts down deployment and deletes the contents of the deployment's
// project directory
func (c *Client) Reset() (*http.Response, error) {
	return c.post("/reset", nil)
}

// Logs get logs of given container
func (c *Client) Logs(container string) (*http.Response, error) {
	reqContent := map[string]string{
		common.Container: container,
	}

	return c.get("/logs", reqContent)
}

// LogsWebSocket opens a websocket connection to given container's logs
func (c *Client) LogsWebSocket(container string) (SocketReader, error) {
	host, err := url.Parse("https://" + c.RemoteVPS.GetIPAndPort())
	if err != nil {
		return nil, err
	}

	// Set up request
	url := &url.URL{Scheme: "wss", Host: host.Host, Path: "/logs"}
	encodeQuery(url, map[string]string{
		common.Container: container,
		common.Stream:    "true",
	})

	// Set up authorization
	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.Daemon.Token)

	// Attempt websocket connection
	socket, resp, err := buildWebSocketDialer().Dial(url.String(), header)
	if err == websocket.ErrBadHandshake {
		return nil, fmt.Errorf("websocket handshake failed with status %d", resp.StatusCode)
	}
	return socket, nil
}

// AddUser adds an authorized user for access to Inertia Web
func (c *Client) AddUser(username, password string, admin bool) (*http.Response, error) {
	reqContent := &common.UserRequest{
		Username: username,
		Password: password,
		Admin:    admin,
	}
	return c.post("/user/adduser", reqContent)
}

// RemoveUser prevents a user from accessing Inertia Web
func (c *Client) RemoveUser(username string) (*http.Response, error) {
	reqContent := &common.UserRequest{Username: username}
	return c.post("/user/removeuser", reqContent)
}

// ResetUsers resets all users on the remote.
func (c *Client) ResetUsers() (*http.Response, error) {
	return c.post("/user/resetusers", nil)
}

// ListUsers lists all users on the remote.
func (c *Client) ListUsers() (*http.Response, error) {
	return c.get("/user/listusers", nil)
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

	client := buildHTTPSClient()
	return client.Do(req)
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

	client := buildHTTPSClient()
	return client.Do(req)
}

func (c *Client) buildRequest(method string, endpoint string, payload io.Reader) (*http.Request, error) {
	// Assemble URL
	url, err := url.Parse("https://" + c.RemoteVPS.GetIPAndPort())
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, endpoint)
	urlString := url.String()

	// Assemble request
	req, err := http.NewRequest(method, urlString, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Daemon.Token)

	return req, nil
}

func buildHTTPSClient() *http.Client {
	return &http.Client{Transport: &http.Transport{
		// Our certificates are self-signed, so will raise
		// a warning - currently, we ask our client to ignore
		// this warning.
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}
}

func buildWebSocketDialer() *websocket.Dialer {
	return &websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func encodeQuery(url *url.URL, queries map[string]string) {
	q := url.Query()
	for k, v := range queries {
		q.Add(k, v)
	}
	url.RawQuery = q.Encode()
}
