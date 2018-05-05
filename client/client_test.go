package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	mockRemote := &RemoteVPS{
		User: "",
		IP:   url,
		PEM:  "",
		Daemon: &DaemonConfig{
			Port:   port,
			Secret: "arjan",
			Token:  fakeAuth,
		},
	}

	return &Client{
		RemoteVPS: mockRemote,
		project:   "test_project",
	}
}

func getIntegrationClient() *Client {
	remote := &RemoteVPS{
		IP:   "127.0.0.1",
		PEM:  "../test/keys/id_rsa",
		User: "root",
		Daemon: &DaemonConfig{
			Port: "4303",
		},
	}
	travis := os.Getenv("TRAVIS")
	if travis != "" {
		remote.Daemon.SSHPort = "69"
	} else {
		remote.Daemon.SSHPort = "22"
	}
	return &Client{
		RemoteVPS: remote,
	}
}

func TestInstallDocker(t *testing.T) {
	remote := getIntegrationClient()
	script, err := ioutil.ReadFile("bootstrap/docker.sh")
	assert.Nil(t, err)

	// Make sure the right command is run.
	session := mockSSHRunner{c: remote}
	remote.installDocker(&session)
	assert.Equal(t, string(script), session.Calls[0])
}

func TestDaemonUp(t *testing.T) {
	remote := getIntegrationClient()
	script, err := ioutil.ReadFile("bootstrap/daemon-up.sh")
	assert.Nil(t, err)
	actualCommand := fmt.Sprintf(string(script), "latest", "4303", "0.0.0.0")

	// Make sure the right command is run.
	session := mockSSHRunner{c: remote}

	// Make sure the right command is run.
	err = remote.DaemonUp(&session, "latest", "0.0.0.0", "4303")
	assert.Nil(t, err)
	println(actualCommand)
	assert.Equal(t, actualCommand, session.Calls[0])
}

func TestKeyGen(t *testing.T) {
	remote := getIntegrationClient()
	script, err := ioutil.ReadFile("bootstrap/token.sh")
	assert.Nil(t, err)
	tokenScript := fmt.Sprintf(string(script), "test")

	// Make sure the right command is run.
	session := mockSSHRunner{c: remote}

	// Make sure the right command is run.
	_, err = remote.getDaemonAPIToken(&session, "test")
	assert.Nil(t, err)
	assert.Equal(t, session.Calls[0], tokenScript)
}

func TestBootstrap(t *testing.T) {
	remote := getIntegrationClient()
	dockerScript, err := ioutil.ReadFile("bootstrap/docker.sh")
	assert.Nil(t, err)

	keyScript, err := ioutil.ReadFile("bootstrap/keygen.sh")
	assert.Nil(t, err)

	script, err := ioutil.ReadFile("bootstrap/token.sh")
	assert.Nil(t, err)
	tokenScript := fmt.Sprintf(string(script), "test")

	script, err = ioutil.ReadFile("bootstrap/daemon-up.sh")
	assert.Nil(t, err)
	daemonScript := fmt.Sprintf(string(script), "test", "4303", "127.0.0.1")

	session := mockSSHRunner{c: remote}
	err = remote.BootstrapRemote(&session)
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

	remote := getIntegrationClient()
	session := &SSHRunner{c: remote}
	err := remote.BootstrapRemote(session)
	assert.Nil(t, err)

	// Daemon setup takes a bit of time - do a crude wait
	time.Sleep(3 * time.Second)

	// Check if daemon is online following bootstrap
	host := "https://" + remote.GetIPAndPort()
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
		var upReq common.DaemonRequest
		err = json.Unmarshal(body, &upReq)
		assert.Nil(t, err)
		assert.Equal(t, "myremote.git", upReq.GitOptions.RemoteURL)
		assert.Equal(t, "arjan", upReq.Secret)
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
		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err)
		defer req.Body.Close()
		var upReq common.DaemonRequest
		err = json.Unmarshal(body, &upReq)
		assert.Nil(t, err)
		assert.Equal(t, "docker-compose", upReq.Container)
		assert.Equal(t, true, upReq.Stream)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := getMockClient(testServer)
	resp, err := d.Logs(true, "docker-compose")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
