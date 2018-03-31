package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/daemon/inertia/project"
)

func TestStatusHandlerBuildInProgress(t *testing.T) {
	defer func() { deployment = nil }()
	// Set up condition
	deployment = &FakeDeployment{
		GetStatusFunc: func(*docker.Client) (*project.DeploymentStatus, error) {
			return &project.DeploymentStatus{
				Branch:               "wow",
				CommitHash:           "abcde",
				CommitMessage:        "",
				Containers:           []string{},
				BuildContainerActive: true,
			}, nil
		},
	}

	// Assmble request
	req, err := http.NewRequest("POST", "/status", nil)
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.Contains(t, recorder.Body.String(), msgBuildInProgress)
}

func TestStatusHandlerNoContainers(t *testing.T) {
	defer func() { deployment = nil }()
	// Set up condition
	deployment = &FakeDeployment{
		GetStatusFunc: func(*docker.Client) (*project.DeploymentStatus, error) {
			return &project.DeploymentStatus{
				Branch:               "wow",
				CommitHash:           "abcde",
				CommitMessage:        "",
				Containers:           []string{},
				BuildContainerActive: false,
			}, nil
		},
	}

	// Assmble request
	req, err := http.NewRequest("POST", "/status", nil)
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.Contains(t, recorder.Body.String(), msgNoContainersActive)
}

func TestStatusHandlerActiveContainers(t *testing.T) {
	defer func() { deployment = nil }()
	// Set up condition
	deployment = &FakeDeployment{
		GetStatusFunc: func(*docker.Client) (*project.DeploymentStatus, error) {
			return &project.DeploymentStatus{
				Branch:               "wow",
				CommitHash:           "abcde",
				CommitMessage:        "",
				Containers:           []string{"mycontainer_1", "yourcontainer_2"},
				BuildContainerActive: false,
			}, nil
		},
	}

	// Assmble request
	req, err := http.NewRequest("POST", "/status", nil)
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.NotContains(t, recorder.Body.String(), msgNoContainersActive)
	assert.NotContains(t, recorder.Body.String(), msgBuildInProgress)
	assert.Contains(t, recorder.Body.String(), "mycontainer_1")
	assert.Contains(t, recorder.Body.String(), "yourcontainer_2")
}

func TestStatusHandlerStatusError(t *testing.T) {
	defer func() { deployment = nil }()
	// Set up condition
	deployment = &FakeDeployment{
		GetStatusFunc: func(*docker.Client) (*project.DeploymentStatus, error) {
			return nil, errors.New("uh oh")
		},
	}

	// Assmble request
	req, err := http.NewRequest("POST", "/status", nil)
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusInternalServerError)
}

func TestStatusHandlerNoDeployment(t *testing.T) {
	// Assmble request
	req, err := http.NewRequest("POST", "/status", nil)
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(statusHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusNotFound)
	assert.Contains(t, recorder.Body.String(), msgNoDeployment)
}
