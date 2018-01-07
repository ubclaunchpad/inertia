package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var (
	fakeAuth = "ubclaunchpad"
)

func getMockDeployment(ts *httptest.Server, s *memory.Storage) (*Deployment, error) {
	wholeURL := strings.Split(ts.URL, ":")
	url := strings.Trim(wholeURL[1], "/")
	port := wholeURL[2]
	mockRemote := &RemoteVPS{
		User: "",
		IP:   url,
		PEM:  "",
		Port: port,
	}
	mockRepo, err := git.Init(s, nil)
	if err != nil {
		return nil, err
	}
	_, err = mockRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"www.myremote.com"},
	})
	if err != nil {
		return nil, err
	}

	return &Deployment{
		RemoteVPS:  mockRemote,
		Repository: mockRepo,
		Auth:       fakeAuth,
	}, nil
}

func TestUp(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

		// Check request body
		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err)
		defer req.Body.Close()
		var upReq UpRequest
		err = json.Unmarshal(body, &upReq)
		assert.Nil(t, err)
		assert.Equal(t, "www.myremote.com", upReq.Repo)

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

	resp, err := d.Up()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDown(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		// Check request method
		assert.Equal(t, "POST", req.Method)

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

func TestReset(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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
