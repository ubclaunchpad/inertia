package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/common"
)

var (
	TestTokenGenerated = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzZXNzaW9uX2lkIjoiIiwidXNlciI6Im1hc3RlciIsImFkbWluIjp0cnVlLCJleHBpcnkiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiJ9.BtWKKZsdqGXDoTZPV6id5PkS0VEVUaLzIangLr2c2XM"
)

func TestTokenHandler(t *testing.T) {
	var testInertiaKeyPath = path.Join(os.Getenv("GOPATH"), "/src/github.com/ubclaunchpad/inertia/test/keys/id_rsa")

	defer os.Setenv("INERTIA_GH_KEY_PATH", "")
	os.Setenv("INERTIA_GH_KEY_PATH", testInertiaKeyPath)

	deployment = &FakeDeployment{
		GetStatusFunc: func(*docker.Client) (common.DeploymentStatus, error) {
			return common.DeploymentStatus{
				Containers: []string{},
			}, nil
		},
	}

	// Assemble request
	req, err := http.NewRequest(http.MethodGet, "/token", nil)
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(tokenHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, TestTokenGenerated, recorder.Body.String())
	assert.Equal(t, http.StatusOK, recorder.Code)
}
