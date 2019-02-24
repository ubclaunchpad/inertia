package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project/mocks"
)

func TestDownHandlerNoDeployment(t *testing.T) {
	var s = &Server{
		deployment: &mocks.FakeDeployer{
			GetStatusStub: func(*docker.Client) (api.DeploymentStatus, error) {
				return api.DeploymentStatus{
					Containers: []string{},
				}, nil
			},
		},
	}

	// Assmble request
	req, err := http.NewRequest("POST", "/down", nil)
	assert.NoError(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(s.downHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusPreconditionFailed)
	assert.Contains(t, recorder.Body.String(), msgNoDeployment)
}
