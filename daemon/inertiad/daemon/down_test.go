package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/mocks"
)

func TestDownHandlerNoDeployment(t *testing.T) {
	var s = &Server{
		deployment: &mocks.FakeDeployer{
			GetStatusStub: func(*docker.Client) (common.DeploymentStatus, error) {
				return common.DeploymentStatus{
					Containers: []string{},
				}, nil
			},
		},
	}

	// Assmble request
	req, err := http.NewRequest("POST", "/down", nil)
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(s.downHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusPreconditionFailed)
	assert.Contains(t, recorder.Body.String(), msgNoDeployment)
}
