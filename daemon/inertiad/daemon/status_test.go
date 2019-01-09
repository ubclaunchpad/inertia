package daemon

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project/mocks"
)

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
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(s.statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
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
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(s.statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
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
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(s.statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.Contains(t, recorder.Body.String(), "mycontainer_1")
	assert.Contains(t, recorder.Body.String(), "yourcontainer_2")
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
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(s.statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusInternalServerError)
}
