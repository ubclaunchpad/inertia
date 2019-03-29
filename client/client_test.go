package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

var (
	fakeAuth = "ubclaunchpad"
)

type testWriter struct{ t *testing.T }

func (l *testWriter) Write(b []byte) (bytes int, e error) { l.t.Log(string(b)); return }

func newMockClient(t *testing.T, ts *httptest.Server) *Client {
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
		debug: true,
		out:   &testWriter{t},
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
	}
}

func TestNewClient(t *testing.T) {
	var c = NewClient(&cfg.Remote{}, Options{})
	assert.NotNil(t, c)
	c.WithDebug(false)
	assert.False(t, c.debug)
	c.WithWriter(os.Stdout)
	assert.Equal(t, os.Stdout, c.out)
}

func TestClient_Up(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "POST", r.Method)

		// Check request body
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()
		var upReq api.UpRequest
		err = json.Unmarshal(body, &upReq)
		assert.NoError(t, err)
		assert.Equal(t, "myremote.git", upReq.GitOptions.RemoteURL)
		assert.Equal(t, "arjan", upReq.WebHookSecret)
		assert.Equal(t, "test_project", upReq.Project)
		assert.Equal(t, "docker-compose", upReq.BuildType)

		// Check correct endpoint called
		assert.Equal(t, "/up", r.URL.Path)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))
		render.Render(w, r, res.MsgOK("uwu"))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer)
	assert.False(t, d.Remote.Daemon.VerifySSL)
	assert.NoError(t, d.Up(context.Background(), UpRequest{"test_project", "myremote.git", cfg.Profile{
		Build: &cfg.Build{
			Type: cfg.DockerCompose,
		},
	}}))
}

func TestClient_UpWithOutput(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
		time.Sleep(10 * time.Millisecond)
		fmt.Fprintln(w, "world")
		time.Sleep(10 * time.Millisecond)
		fmt.Fprintln(w, "chicken rice")
		time.Sleep(10 * time.Millisecond)
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer)
	var buf = &bytes.Buffer{}
	d.out = buf
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// wait before closing the connection to make sure messages arrive
		time.Sleep(1 * time.Second)
		cancel()
	}()
	assert.NoError(t, d.UpWithOutput(ctx, UpRequest{"test_project", "myremote.git", cfg.Profile{
		Build: &cfg.Build{
			Type: cfg.DockerCompose,
		},
	}}))
	assert.Contains(t, buf.String(), "hello\nworld")
	assert.Contains(t, buf.String(), "chicken rice")
}

func TestClient_Prune(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "POST", r.Method)

		// Check correct endpoint called
		endpoint := r.URL.Path
		assert.Equal(t, "/prune", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))
		render.Render(w, r, res.MsgOK("uwu"))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer)
	assert.NoError(t, d.Prune(context.Background()))
}

func TestClient_Down(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "POST", r.Method)

		// Check correct endpoint called
		assert.Equal(t, "/down", r.URL.Path)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

		render.Render(w, r, res.MsgOK("uwu"))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer)
	assert.NoError(t, d.Down(context.Background()))
}

func TestClient_Status(t *testing.T) {
	t.Run("daemon online", func(t *testing.T) {
		testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Check request method
			assert.Equal(t, "GET", r.Method)

			// Check correct endpoint called
			assert.Equal(t, "/status", r.URL.Path)

			// Check auth
			assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

			// return something
			render.Render(w, r, res.MsgOK("status retrieved",
				"status", api.DeploymentStatus{Branch: "amazing_test"}))
		}))
		defer testServer.Close()

		var d = newMockClient(t, testServer)
		status, err := d.Status(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "amazing_test", status.Branch)
	})

	t.Run("daemon offline", func(t *testing.T) {
		var d = newMockClient(t, nil)
		_, err := d.Status(context.Background())
		assert.Contains(t, err.Error(), "appears offline")
	})
}

func TestClient_Reset(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "POST", r.Method)

		// Check correct endpoint called
		assert.Equal(t, "/reset", r.URL.Path)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

		render.Render(w, r, res.MsgOK("uwu"))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer)
	assert.NoError(t, d.Reset(context.Background()))
}

func TestClient_Logs(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "GET", r.Method)

		// Check correct endpoint called
		endpoint := r.URL.Path
		assert.Equal(t, "/logs", endpoint)

		// Check body
		q := r.URL.Query()
		assert.Equal(t, "docker-compose", q.Get(api.Container))
		assert.Equal(t, "10", q.Get(api.Entries))

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

		// return response
		render.Render(w, r, res.MsgOK("logs retrieved",
			"logs", []string{"hello", "world"}))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer)
	logs, err := d.Logs(context.Background(), LogsRequest{"docker-compose", 10})
	assert.NoError(t, err)
	assert.Equal(t, []string{"hello", "world"}, logs)
}

func TestClient_LogsWithOutput(t *testing.T) {
	t.Run("daemon online", func(t *testing.T) {
		testServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// Check request method
			assert.Equal(t, "GET", req.Method)

			// Check correct endpoint called
			endpoint := req.URL.Path
			assert.Equal(t, "/logs", endpoint)

			// Check body
			defer req.Body.Close()
			var q = req.URL.Query()
			assert.Equal(t, "docker-compose", q.Get(api.Container))
			assert.Equal(t, "true", q.Get(api.Stream))
			assert.Equal(t, "10", q.Get(api.Entries))

			// Check auth
			assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))

			// upgrade to websocket and write back
			var socketUpgrader = websocket.Upgrader{}
			socket, err := socketUpgrader.Upgrade(rw, req, nil)
			assert.NoError(t, err)
			assert.NoError(t, socket.WriteMessage(
				websocket.TextMessage, []byte("hello world")))
		}))
		defer testServer.Close()

		var d = newMockClient(t, testServer)
		var buf = &bytes.Buffer{}
		d.out = buf
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			// wait before closing the connection to make sure message arrives
			time.Sleep(1 * time.Second)
			cancel()
		}()
		assert.NoError(t, d.LogsWithOutput(ctx, LogsRequest{"docker-compose", 10}))
		assert.Contains(t, buf.String(), "hello world")
	})

	t.Run("daemon offline", func(t *testing.T) {
		testServer := httptest.NewTLSServer(nil)
		// close the server to test error
		testServer.Close()

		var d = newMockClient(t, testServer)
		var err = d.LogsWithOutput(context.Background(), LogsRequest{"docker-compose", 10})
		assert.Error(t, err)
		assert.True(t,
			strings.Contains(err.Error(), "connect: connection refused") ||
				strings.Contains(err.Error(), "connectex: No connection could be made"))
	})
}

func TestClient_UpdateEnv(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "POST", r.Method)

		// Check correct endpoint called
		assert.Equal(t, "/env", r.URL.Path)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

		render.Render(w, r, res.MsgOK("uwu"))
	}))
	defer testServer.Close()

	d := newMockClient(t, testServer)
	assert.NoError(t, d.UpdateEnv(context.Background(), "", "", false, false))
}

func TestClient_ListEnv(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "GET", r.Method)

		// Check correct endpoint called
		endpoint := r.URL.Path
		assert.Equal(t, "/env", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))
		render.Render(w, r, res.Msg("configured environment variables retrieved", http.StatusOK,
			"variables", []string{"hello", "world"}))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer)
	envs, err := d.ListEnv(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"hello", "world"}, envs)
}

func TestClient_Token(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, http.MethodGet, r.Method)

		// Check correct endpoint called
		endpoint := r.URL.Path
		assert.Equal(t, "/token", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

		render.Render(w, r, res.MsgOK("token generated",
			"token", "hello-world"))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer)
	token, err := d.Token(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "hello-world", token)
}
