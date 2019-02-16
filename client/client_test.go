package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/cfg"
)

var (
	fakeAuth = "ubclaunchpad"
)

func newMockClient(ts *httptest.Server) *Client {
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
		out:       os.Stdout,
		project:   "test_project",
	}
}

func newMockSSHClient(mockRunner *mockSSHRunner) *Client {
	remote := &cfg.RemoteVPS{
		IP:      "127.0.0.1",
		PEM:     "../test/keys/id_rsa",
		User:    "root",
		SSHPort: "69",
		Daemon: &cfg.DaemonConfig{
			Port: "4303",
		},
	}
	mockRunner.r = remote
	return &Client{
		version:   "test",
		RemoteVPS: remote,
		out:       os.Stdout,
		SSH:       mockRunner,
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

	_, found := NewClient("tst", "", config)
	assert.False(t, found)

	cli, found := NewClient("test", "", config)
	assert.True(t, found)
	assert.Equal(t, "/some/pem/file", cli.RemoteVPS.PEM)
	assert.Equal(t, "test", cli.version)
	assert.Equal(t, "robert-writes-bad-code", cli.project)
	assert.False(t, cli.verifySSL)
}

func TestInstallDocker(t *testing.T) {
	session := &mockSSHRunner{}
	client := newMockSSHClient(session)
	script, err := ioutil.ReadFile("scripts/docker.sh")
	assert.NoError(t, err)

	// Make sure the right command is run.
	err = client.installDocker(session)
	assert.NoError(t, err)
	assert.Equal(t, string(script), session.Calls[0])
}

func TestDaemonUp(t *testing.T) {
	session := &mockSSHRunner{}
	client := newMockSSHClient(session)
	client.version = "latest"
	client.IP = "0.0.0.0"
	client.Daemon.Port = "4303"
	script, err := ioutil.ReadFile("scripts/daemon-up.sh")
	assert.NoError(t, err)
	actualCommand := fmt.Sprintf(string(script), "latest", "4303", "0.0.0.0")

	// Make sure the right command is run.
	err = client.DaemonUp("latest")
	assert.NoError(t, err)
	println(actualCommand)
	assert.Equal(t, actualCommand, session.Calls[0])
}

func TestDaemonDown(t *testing.T) {
	session := &mockSSHRunner{}
	client := newMockSSHClient(session)
	client.version = "latest"
	client.IP = "0.0.0.0"
	client.Daemon.Port = "4303"
	script, err := ioutil.ReadFile("scripts/daemon-down.sh")
	assert.NoError(t, err)
	actualCommand := fmt.Sprintf(string(script))

	// Make sure the right command is run.
	err = client.DaemonDown()
	assert.NoError(t, err)
	println(actualCommand)
	assert.Equal(t, actualCommand, session.Calls[0])
}

func TestDaemonDown(t *testing.T) {
	session := &mockSSHRunner{}
	client := newMockSSHClient(session)
	client.version = "latest"
	client.IP = "0.0.0.0"
	client.Daemon.Port = "4303"
	script, err := ioutil.ReadFile("scripts/daemon-down.sh")
	assert.Nil(t, err)
	actualCommand := fmt.Sprintf(string(script))

	// Make sure the right command is run.
	err = client.DaemonDown()
	assert.Nil(t, err)
	println(actualCommand)
	assert.Equal(t, actualCommand, session.Calls[0])
}

func TestKeyGen(t *testing.T) {
	session := &mockSSHRunner{}
	remote := newMockSSHClient(session)
	script, err := ioutil.ReadFile("scripts/token.sh")
	assert.NoError(t, err)
	tokenScript := fmt.Sprintf(string(script), "test")

	// Make sure the right command is run.
	_, err = remote.getDaemonAPIToken(session, "test")
	assert.NoError(t, err)
	assert.Equal(t, session.Calls[0], tokenScript)
}

func TestBootstrap(t *testing.T) {
	session := &mockSSHRunner{}
	client := newMockSSHClient(session)
	assert.False(t, client.verifySSL)

	dockerScript, err := ioutil.ReadFile("scripts/docker.sh")
	assert.NoError(t, err)

	keyScript, err := ioutil.ReadFile("scripts/keygen.sh")
	assert.NoError(t, err)

	script, err := ioutil.ReadFile("scripts/token.sh")
	assert.NoError(t, err)
	tokenScript := fmt.Sprintf(string(script), "test")

	script, err = ioutil.ReadFile("scripts/daemon-up.sh")
	assert.NoError(t, err)
	daemonScript := fmt.Sprintf(string(script), "test", "4303", "127.0.0.1")

	err = client.BootstrapRemote("ubclaunchpad/inertia")
	assert.NoError(t, err)

	// Make sure all commands are formatted correctly
	assert.Equal(t, string(dockerScript), session.Calls[0])
	assert.Equal(t, string(keyScript), session.Calls[1])
	assert.Equal(t, daemonScript, session.Calls[2])
	assert.Equal(t, tokenScript, session.Calls[3])
}

func TestUp(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check request body
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)
		defer req.Body.Close()
		var upReq api.UpRequest
		err = json.Unmarshal(body, &upReq)
		assert.NoError(t, err)
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

	d := newMockClient(testServer)
	assert.False(t, d.verifySSL)
	resp, err := d.Up("myremote.git", "docker-compose", false)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPrune(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/prune", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.Prune()
	assert.NoError(t, err)
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

	d := newMockClient(testServer)
	resp, err := d.Down()
	assert.NoError(t, err)
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

	d := newMockClient(testServer)
	resp, err := d.Status()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestStatusFail(t *testing.T) {
	d := newMockClient(nil)
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

	d := newMockClient(testServer)
	resp, err := d.Reset()
	assert.NoError(t, err)
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
		assert.Equal(t, "docker-compose", q.Get(api.Container))
		assert.Equal(t, "10", q.Get(api.Entries))

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.Logs("docker-compose", 10)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLogsWebsocket(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Check request method
		assert.Equal(t, "GET", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/logs", endpoint)

		// Check body
		defer req.Body.Close()
		q := req.URL.Query()
		assert.Equal(t, "docker-compose", q.Get(api.Container))
		assert.Equal(t, "true", q.Get(api.Stream))
		assert.Equal(t, "10", q.Get(api.Entries))

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))

		socketUpgrader := websocket.Upgrader{}
		socket, err := socketUpgrader.Upgrade(rw, req, nil)
		assert.NoError(t, err)

		err = socket.WriteMessage(websocket.TextMessage, []byte("hello world"))
		assert.NoError(t, err)
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.LogsWebSocket("docker-compose", 10)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	_, m, err := resp.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world"), m)
}

func TestLogsWebsocketNoDaemon(t *testing.T) {
	testServer := httptest.NewTLSServer(nil)
	// close the server to test error
	testServer.Close()

	d := newMockClient(testServer)
	_, err := d.LogsWebSocket("docker-compose", 10)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "connect: connection refused") || strings.Contains(err.Error(), "connectex: No connection could be made"))
}

func TestUpdateEnv(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/env", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.UpdateEnv("", "", false, false)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestListEnv(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "GET", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/env", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.ListEnv()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAddUser(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/user/add", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.AddUser("", "", false)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveUser(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/user/remove", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.RemoveUser("")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestResetUser(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/user/reset", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.ResetUsers()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestListUsers(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "GET", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/user/list", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.ListUsers()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestToken(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, http.MethodGet, req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/token", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.Token()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLogIn(t *testing.T) {
	username := "testguy"
	password := "SomeKindo23asdfpassword"
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, http.MethodPost, req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/user/login", endpoint)

		// Check auth
		defer req.Body.Close()
		body, err := ioutil.ReadAll(req.Body)
		assert.Equal(t, nil, err)
		var userReq api.UserRequest
		assert.Equal(t, nil, json.Unmarshal(body, &userReq))
		assert.Equal(t, userReq.Username, username)
		assert.Equal(t, userReq.Password, password)
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.LogIn(username, password, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestEnableTotp(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/user/totp/enable", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.EnableTotp("", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDisableTotp(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/user/totp/disable", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.DisableTotp()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestEnableTotp(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/user/totp/enable", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.EnableTotp("", "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDisableTotp(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/user/totp/disable", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	resp, err := d.DisableTotp()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetSSLVerification(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check correct endpoint called
		endpoint := req.URL.Path
		assert.Equal(t, "/user/totp/disable", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	d := newMockClient(testServer)
	d.SetSSLVerification(true)
	assert.Equal(t, d.verifySSL, true)
}
