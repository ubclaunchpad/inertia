package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/common"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var (
	fakeAuth = "ubclaunchpad"
)

func getMockDeployment(ts *httptest.Server, s *memory.Storage) (*Deployment, error) {
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
		},
	}
	mockRepo, err := git.Init(s, nil)
	if err != nil {
		return nil, err
	}
	_, err = mockRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"myremote"},
	})
	if err != nil {
		return nil, err
	}

	return &Deployment{
		RemoteVPS:  mockRemote,
		Repository: mockRepo,
		Auth:       fakeAuth,
		Project:    "test_project",
	}, nil
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

	memory := memory.NewStorage()
	defer func() { memory = nil }()

	d, err := getMockDeployment(testServer, memory)
	assert.Nil(t, err)

	resp, err := d.Up("docker-compose", false)
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

	memory := memory.NewStorage()
	defer func() { memory = nil }()

	d, err := getMockDeployment(testServer, memory)
	assert.Nil(t, err)

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

	memory := memory.NewStorage()
	defer func() { memory = nil }()

	d, err := getMockDeployment(testServer, memory)
	assert.Nil(t, err)

	resp, err := d.Status()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestStatusFail(t *testing.T) {
	memory := memory.NewStorage()
	defer func() { memory = nil }()

	d, err := getMockDeployment(nil, memory)
	assert.Nil(t, err)

	_, err = d.Status()
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

	memory := memory.NewStorage()
	defer func() { memory = nil }()

	d, err := getMockDeployment(testServer, memory)
	assert.Nil(t, err)

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
		assert.Equal(t, "true", q.Get(common.Stream))

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, req.Header.Get("Authorization"))
	}))
	defer testServer.Close()

	memory := memory.NewStorage()
	defer func() { memory = nil }()

	d, err := getMockDeployment(testServer, memory)
	assert.Nil(t, err)

	resp, err := d.Logs(true, "docker-compose")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
