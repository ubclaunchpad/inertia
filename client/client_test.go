package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/common"
)

var (
	fakeAuth = "ubclaunchpad"
)

func getMockClient(ts *httptest.Server) *Client {
	var (
		url  string
		port string
	)
	if ts != nil {
		wholeURL := strings.Split(ts.URL, ":")
		url = strings.Trim(wholeURL[1], "/")
		port = wholeURL[2]
	} else {
		url = "0.0.0.0"
		port = "8080"
	}

	mockRemote := &cfg.RemoteVPS{
		User: "",
		IP:   url,
		PEM:  "",
		Daemon: &cfg.DaemonConfig{
			Port:          port,
			WebHookSecret: "arjan",
			Token:         fakeAuth,
		},
	}

	return &Client{
		RemoteVPS: mockRemote,
		project:   "test_project",
	}
}

func getIntegrationClient(mockRunner *mockSSHRunner) *Client {
	remote := &cfg.RemoteVPS{
		IP:      "127.0.0.1",
		PEM:     "../test/keys/id_rsa",
		User:    "root",
		SSHPort: "69",
		Daemon: &cfg.DaemonConfig{
			Port: "4303",
		},
	}
	if mockRunner != nil {
		mockRunner.r = remote
		return &Client{
			version:   "test",
			RemoteVPS: remote,
			sshRunner: mockRunner,
		}
	}
	return &Client{
		version:   "test",
		RemoteVPS: remote,
		sshRunner: NewSSHRunner(remote),
	}
}

func TestGetNewClient(t *testing.T) {
	config := &cfg.Config{
		Version: "test",
		Project: "robert-writes-bad-code",
		Remotes: make(map[string]*cfg.RemoteVPS),
	}
	testRemote := &cfg.RemoteVPS{
		Name:    "test",
		IP:      "12343",
		User:    "bobheadxi",
		PEM:     "/some/pem/file",
		SSHPort: "22",
		Daemon: &cfg.DaemonConfig{
			Port: "8080",
		},
	}
	config.AddRemote(testRemote)

	_, found := NewClient("tst", config)
	assert.False(t, found)

	cli, found := NewClient("test", config)
	assert.True(t, found)
	assert.Equal(t, "/some/pem/file", cli.RemoteVPS.PEM)
	assert.Equal(t, "test", cli.version)
	assert.Equal(t, "robert-writes-bad-code", cli.project)
	assert.False(t, cli.verifySSL)
}

func TestInstallDocker(t *testing.T) {
	session := &mockSSHRunner{}
	client := getIntegrationClient(session)
	script, err := ioutil.ReadFile("scripts/docker.sh")
	assert.Nil(t, err)

	// Make sure the right command is run.
	err = client.installDocker(session)
	assert.Nil(t, err)
	assert.Equal(t, string(script), session.Calls[0])
}

func TestDaemonUp(t *testing.T) {
	session := &mockSSHRunner{}
	client := getIntegrationClient(session)
	script, err := ioutil.ReadFile("scripts/daemon-up.sh")
	assert.Nil(t, err)
	actualCommand := fmt.Sprintf(string(script), "latest", "4303", "0.0.0.0")

	// Make sure the right command is run.
	err = client.DaemonUp("latest", "0.0.0.0", "4303")
	assert.Nil(t, err)
	println(actualCommand)
	assert.Equal(t, actualCommand, session.Calls[0])
}

func TestKeyGen(t *testing.T) {
	session := &mockSSHRunner{}
	remote := getIntegrationClient(session)
	script, err := ioutil.ReadFile("scripts/token.sh")
	assert.Nil(t, err)
	tokenScript := fmt.Sprintf(string(script), "test")

	// Make sure the right command is run.
	_, err = remote.getDaemonAPIToken(session, "test")
	assert.Nil(t, err)
	assert.Equal(t, session.Calls[0], tokenScript)
}

func TestBootstrap(t *testing.T) {
	session := &mockSSHRunner{}
	client := getIntegrationClient(session)
	assert.False(t, client.verifySSL)

	dockerScript, err := ioutil.ReadFile("scripts/docker.sh")
	assert.Nil(t, err)

	keyScript, err := ioutil.ReadFile("scripts/keygen.sh")
	assert.Nil(t, err)

	script, err := ioutil.ReadFile("scripts/token.sh")
	assert.Nil(t, err)
	tokenScript := fmt.Sprintf(string(script), "test")

	script, err = ioutil.ReadFile("scripts/daemon-up.sh")
	assert.Nil(t, err)
	daemonScript := fmt.Sprintf(string(script), "test", "4303", "127.0.0.1")

	err = client.BootstrapRemote("ubclaunchpad/inertia")
	assert.Nil(t, err)

	// Make sure all commands are formatted correctly
	assert.Equal(t, string(dockerScript), session.Calls[0])
	assert.Equal(t, string(keyScript), session.Calls[1])
	assert.Equal(t, daemonScript, session.Calls[2])
	assert.Equal(t, tokenScript, session.Calls[3])
}

func TestBootstrapIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cli := getIntegrationClient(nil)
	err := cli.BootstrapRemote("")
	assert.Nil(t, err)

	// Daemon setup takes a bit of time - do a crude wait
	time.Sleep(3 * time.Second)

	// Check if daemon is online following bootstrap
	host := "https://" + cli.GetIPAndPort()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(host)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
}

func TestUp(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check request body
		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err)
		defer req.Body.Close()
		var upReq common.UpRequest
		err = json.Unmarshal(body, &upReq)
		assert.Nil(t, err)
		assert.Equal(t, "myremote.git", upReq.GitOptions.RemoteURL)
		assert.Equal(t, "arjan", upReq.WebHookSecret)
		assert.Equal(t, "test_project", upReq.Project)
		assert.Equal(t, "docker-compose", upReq.BuildType)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/up", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := getMockClient(testServer)
	assert.False(t, d.verifySSL)
	resp, err := d.Up("myremote.git", "docker-compose", false)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDown(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/down", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := getMockClient(testServer)
	resp, err := d.Down()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestStatus(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "GET", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/status", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := getMockClient(testServer)
	resp, err := d.Status()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestStatusFail(t *testing.T) {
	d := getMockClient(nil)
	_, err := d.Status()
	assert.Contains(t, err.Error(), "appears offline")
}

func TestReset(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/reset", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := getMockClient(testServer)
	resp, err := d.Reset()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLogs(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "GET", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/logs", endpoint)

		// Check body
		defer req.Body.Close()
		q := req.URL.Query()
		assert.Equal(t, "docker-compose", q.Get(common.Container))

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := getMockClient(testServer)
	resp, err := d.Logs("docker-compose")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
