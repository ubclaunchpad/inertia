package client

import (
	"encoding/json"
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
	"github.com/ubclaunchpad/inertia/client/runner/mocks"
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

	return &Client{
		Remote: &cfg.Remote{
			IP: url,
			SSH: &cfg.SSH{
				User:         "",
				IdentityFile: "",
			},
			Daemon: &cfg.Daemon{
				Port:          port,
				WebHookSecret: "arjan",
				Token:         fakeAuth,
			},
		},
		out: os.Stdout,
	}
}

func newMockSSHClient(m *mocks.FakeSSHSession) *Client {
	return &Client{
		Remote: &cfg.Remote{
			Version: "test",
			IP:      "127.0.0.1",
			SSH: &cfg.SSH{
				IdentityFile: "../test/keys/id_rsa",
				User:         "root",
				SSHPort:      "69",
			},
			Daemon: &cfg.Daemon{
				Port: "4303",
			},
		},
		out: os.Stdout,
		ssh: m,
	}
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

	var d = newMockClient(testServer)
	assert.False(t, d.Remote.Daemon.VerifySSL)
	resp, err := d.Up("test_project", "myremote.git", cfg.Profile{
		Build: &cfg.Build{
			Type: cfg.DockerCompose,
		},
	}, false)
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
