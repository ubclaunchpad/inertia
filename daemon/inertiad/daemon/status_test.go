package daemon

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project/mocks"
)

func readBadge(body io.Reader) (*shieldsIOData, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	bytes := buf.Bytes()
	var data shieldsIOData
	return &data, json.Unmarshal(bytes, &data)
}

func TestStatusHandlerBuildInProgress(t *testing.T) {
	var s = &Server{
		deployment: &mocks.FakeDeployer{
			GetStatusStub: func(*docker.Client) (api.DeploymentStatus, error) {
				return api.DeploymentStatus{
					Branch:               "wow",
					CommitHash:           "abcde",
					CommitMessage:        "",
					Containers:           []string{},
					BuildContainerActive: true,
				}, nil
			},
		},
	}

	// Assmble request
	req, err := http.NewRequest("GET", "/status", nil)
	assert.NoError(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(s.statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)

	// check badge
	recorder = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/status?badge=true", nil)
	assert.NoError(t, err)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	badge, err := readBadge(recorder.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, "deploying", badge.Message)
	assert.False(t, badge.IsError)
}

func TestStatusHandlerNoContainers(t *testing.T) {
	var s = &Server{
		deployment: &mocks.FakeDeployer{
			GetStatusStub: func(*docker.Client) (api.DeploymentStatus, error) {
				return api.DeploymentStatus{
					Branch:               "wow",
					CommitHash:           "abcde",
					CommitMessage:        "",
					Containers:           []string{},
					BuildContainerActive: false,
				}, nil
			},
		},
	}

	// Assmble request
	req, err := http.NewRequest("GET", "/status", nil)
	assert.NoError(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(s.statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)

	// check badge
	recorder = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/status?badge=true", nil)
	assert.NoError(t, err)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	badge, err := readBadge(recorder.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, "project offline", badge.Message)
	assert.True(t, badge.IsError)
}

func TestStatusHandlerActiveContainers(t *testing.T) {
	var s = &Server{
		deployment: &mocks.FakeDeployer{
			GetStatusStub: func(*docker.Client) (api.DeploymentStatus, error) {
				return api.DeploymentStatus{
					Branch:               "wow",
					CommitHash:           "abcde",
					CommitMessage:        "",
					Containers:           []string{"mycontainer_1", "yourcontainer_2"},
					BuildContainerActive: false,
				}, nil
			},
		},
	}

	// Assmble request
	req, err := http.NewRequest("GET", "/status", nil)
	assert.NoError(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(s.statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.Contains(t, recorder.Body.String(), "mycontainer_1")
	assert.Contains(t, recorder.Body.String(), "yourcontainer_2")

	// check badge
	recorder = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/status?badge=true", nil)
	assert.NoError(t, err)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	badge, err := readBadge(recorder.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, "deployed", badge.Message)
	assert.False(t, badge.IsError)
}

func TestStatusHandlerStatusError(t *testing.T) {
	var s = &Server{
		deployment: &mocks.FakeDeployer{
			GetStatusStub: func(*docker.Client) (api.DeploymentStatus, error) {
				return api.DeploymentStatus{CommitHash: "1234"}, errors.New("uh oh")
			},
		},
	}

	// Assmble request
	req, err := http.NewRequest("GET", "/status", nil)
	assert.NoError(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(s.statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusInternalServerError)

	// check badge
	recorder = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/status?badge=true", nil)
	assert.NoError(t, err)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	badge, err := readBadge(recorder.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, "errored", badge.Message)
	assert.True(t, badge.IsError)
}
