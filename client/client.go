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
	"github.com/ubclaunchpad/inertia/cfg"
	internal "github.com/ubclaunchpad/inertia/client/internal"
	"github.com/ubclaunchpad/inertia/common"
)

// Client manages a deployment
type Client struct {
	*cfg.RemoteVPS
	version       string
	project       string
	buildType     string
	buildFilePath string

	out io.Writer

	sshRunner SSHSession
	verifySSL bool
}

// NewClient sets up a client to communicate to the daemon at
// the given named remote.
func NewClient(remoteName string, config *cfg.Config, out ...io.Writer) (*Client, bool) {
	remote, found := config.GetRemote(remoteName)
	if !found {
		return nil, false
	}

	var writer io.Writer
	if len(out) > 0 {
		writer = out[0]
	} else {
		writer = common.DevNull{}
	}

	return &Client{
		RemoteVPS:     remote,
		version:       config.Version,
		project:       config.Project,
		buildType:     config.BuildType,
		buildFilePath: config.BuildFilePath,
		sshRunner:     NewSSHRunner(remote),

		out: writer,
	}, true
}

// SetSSLVerification toggles whether client should verify all SSL communications.
// This requires a signed certificate to be in use on your daemon.
func (c *Client) SetSSLVerification(verify bool) {
	c.verifySSL = verify
}

// BootstrapRemote configures a remote vps for continuous deployment
// by installing docker, starting the daemon and building a
// public-private key-pair. It outputs configuration information
// for the user.
func (c *Client) BootstrapRemote(repoName string) error {
	fmt.Fprintf(c.out, "Setting up remote %s at %s", c.Name, c.IP)

	fmt.Fprint(c.out, ">> Step 1/4: Installing docker...")
	err := c.installDocker(c.sshRunner)
	if err != nil {
		return err
	}

	fmt.Fprint(c.out, "\n>> Step 2/4: Building deploy key...")
	if err != nil {
		return err
	}
	pub, err := c.keyGen(c.sshRunner)
	if err != nil {
		return err
	}

	// This step needs to run before any other commands that rely on
	// the daemon image, since the daemon is loaded here.
	fmt.Fprint(c.out, "\n>> Step 3/4: Starting daemon...")
	if err != nil {
		return err
	}
	err = c.DaemonUp(c.version, c.IP, c.Daemon.Port)
	if err != nil {
		return err
	}

	fmt.Fprint(c.out, "\n>> Step 4/4: Fetching daemon API token...")
	token, err := c.getDaemonAPIToken(c.sshRunner, c.version)
	if err != nil {
		return err
	}
	c.Daemon.Token = token

	fmt.Fprint(c.out, "\nInertia has been set up and daemon is running on remote!")
	fmt.Fprint(c.out, "You may have to wait briefly for Inertia to set up some dependencies.")
	fmt.Fprintf(c.out, "Use 'inertia %s logs --stream' to check on the daemon's setup progress.\n\n", c.Name)

	fmt.Fprint(c.out, "=============================\n")

	// Output deploy key to user.
	fmt.Fprintf(c.out, ">> GitHub Deploy Key (add to https://www.github.com/%s/settings/keys/new): ", repoName)
	fmt.Fprint(c.out, pub.String())

	// Output Webhook url to user.
	fmt.Fprintf(c.out, ">> GitHub WebHook URL (add to https://www.github.com/%s/settings/hooks/new): ", repoName)
	fmt.Fprintf(c.out, "WebHook Address:  https://%s:%s/webhook", c.IP, c.Daemon.Port)
	fmt.Fprint(c.out, "WebHook Secret:   "+c.Daemon.WebHookSecret)
	fmt.Fprint(c.out, `Note that you will have to disable SSH verification in your webhook
settings - Inertia uses self-signed certificates that GitHub won't
be able to verify.`+"\n")

	fmt.Fprint(c.out, `Inertia daemon successfully deployed! Add your webhook url and deploy
key to enable continuous deployment.`)
	fmt.Fprintf(c.out, "Then run 'inertia %s up' to deploy your application.\n", c.Name)

	return nil
}

// DaemonUp brings the daemon up on the remote instance.
func (c *Client) DaemonUp(daemonVersion, host, daemonPort string) error {
	scriptBytes, err := internal.Asset("client/scripts/daemon-up.sh")
	if err != nil {
		return err
	}

	// Run inertia daemon.
	daemonCmdStr := fmt.Sprintf(string(scriptBytes), daemonVersion, daemonPort, host)
	return c.sshRunner.RunStream(daemonCmdStr, false)
}

// DaemonDown brings the daemon down on the remote instance
func (c *Client) DaemonDown() error {
	scriptBytes, err := internal.Asset("client/scripts/daemon-down.sh")
	if err != nil {
		return err
	}

	_, stderr, err := c.sshRunner.Run(string(scriptBytes))
	if err != nil {
		return fmt.Errorf("daemon shutdown failed: %s: %s", err.Error(), stderr.String())
	}

	return nil
}

// installDocker installs docker on a remote vps.
func (c *Client) installDocker(session SSHSession) error {
	installDockerSh, err := internal.Asset("client/scripts/docker.sh")
	if err != nil {
		return err
	}

	// Install docker.
	cmdStr := string(installDockerSh)
	_, stderr, err := session.Run(cmdStr)
	if err != nil {
		return fmt.Errorf("docker installation: %s: %s", err.Error(), stderr.String())
	}

	return nil
}

// keyGen creates a public-private key-pair on the remote vps
// and returns the public key.
func (c *Client) keyGen(session SSHSession) (*bytes.Buffer, error) {
	scriptBytes, err := internal.Asset("client/scripts/keygen.sh")
	if err != nil {
		return nil, err
	}

	// Create deploy key.
	result, stderr, err := session.Run(string(scriptBytes))

	if err != nil {
		return nil, fmt.Errorf("key generation failed: %s: %s", err.Error(), stderr.String())
	}

	return result, nil
}

// getDaemonAPIToken returns the daemon API token for RESTful access
// to the daemon.
func (c *Client) getDaemonAPIToken(session SSHSession, daemonVersion string) (string, error) {
	scriptBytes, err := internal.Asset("client/scripts/token.sh")
	if err != nil {
		return "", err
	}
	daemonCmdStr := fmt.Sprintf(string(scriptBytes), daemonVersion)

	stdout, stderr, err := session.Run(daemonCmdStr)
	if err != nil {
		return "", fmt.Errorf("api token generation failed: %s: %s", err.Error(), stderr.String())
	}

	// There may be a newline, remove it.
	return strings.TrimSuffix(stdout.String(), "\n"), nil
}

// Up brings the project up on the remote VPS instance specified
// in the deployment object.
func (c *Client) Up(gitRemoteURL, buildType string, stream bool) (*http.Response, error) {
	if buildType == "" {
		buildType = c.buildType
	}

	return c.post("/up", &common.UpRequest{
		Stream:        stream,
		Project:       c.project,
		BuildType:     buildType,
		WebHookSecret: c.RemoteVPS.Daemon.WebHookSecret,
		BuildFilePath: c.buildFilePath,
		GitOptions: &common.GitOptions{
			RemoteURL: common.GetSSHRemoteURL(gitRemoteURL),
			Branch:    c.Branch,
		},
	})
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
	socket, resp, err := buildWebSocketDialer(c.verifySSL).Dial(url.String(), header)
	if err == websocket.ErrBadHandshake {
		return nil, fmt.Errorf("websocket handshake failed with status %d", resp.StatusCode)
	}
	return socket, nil
}

// UpdateEnv updates environment variable
func (c *Client) UpdateEnv(name, value string, encrypt, remove bool) (*http.Response, error) {
	return c.post("/env", common.EnvRequest{
		Name: name, Value: value, Encrypt: encrypt, Remove: remove,
	})
}

// ListEnv lists environment variables currently set on remote
func (c *Client) ListEnv() (*http.Response, error) {
	return c.get("/env", nil)
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

	client := buildHTTPSClient(c.verifySSL)
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

	client := buildHTTPSClient(c.verifySSL)
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

func buildHTTPSClient(verify bool) *http.Client {
	return &http.Client{Transport: &http.Transport{
		// Our certificates are self-signed, so will raise
		// a warning - currently, we ask our client to ignore
		// this warning.
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !verify,
		},
	}}
}

func buildWebSocketDialer(verify bool) *websocket.Dialer {
	return &websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !verify,
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
