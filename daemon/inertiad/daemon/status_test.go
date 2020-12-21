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
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(body); err != nil {
		return nil, err
	}
	var data shieldsIOData
	return &data, json.Unmarshal(buf.Bytes(), &data)
}

func readStatus(t *testing.T, body io.Reader) (*api.DeploymentStatusWithVersions, error) {
	var data api.DeploymentStatusWithVersions
	base, err := api.Unmarshal(body, api.KV{Key: "status", Value: &data})
	if err != nil {
		return nil, err
	}
	t.Log(base.Message)
	return &data, nil
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

func TestStatusHandlerNotUpToDate(t *testing.T) {
	const someOldVersion = "v0.6.0" // outdated version
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
		version: someOldVersion,
	}

	// Assemble request
	req, err := http.NewRequest(http.MethodGet, "/status", nil)
	assert.NoError(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(s.statusHandler)

	handler.ServeHTTP(recorder, req)
	gotStat, err := readStatus(t, recorder.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, recorder.Code, http.StatusOK)

	// Check status
	assert.Equal(t, gotStat.InertiaVersion, someOldVersion)
	assert.NotNil(t, gotStat.NewVersionAvailable, "new version should be available")
	assert.NotEqual(t, gotStat.NewVersionAvailable, s.version)
}
