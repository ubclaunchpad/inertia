package daemon

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenHandler(t *testing.T) {
	var (
		generatedTestToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzZXNzaW9uX2lkIjoiIiwidXNlciI6Im1hc3RlciIsImFkbWluIjp0cnVlLCJleHBpcnkiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiJ9.BtWKKZsdqGXDoTZPV6id5PkS0VEVUaLzIangLr2c2XM"
		testInertiaKeyPath = path.Join(os.Getenv("GOPATH"), "/src/github.com/ubclaunchpad/inertia/test/keys/id_rsa")
	)
	os.Setenv("INERTIA_GH_KEY_PATH", testInertiaKeyPath)
	defer os.Setenv("INERTIA_GH_KEY_PATH", "")

	// Assemble request
	req, err := http.NewRequest(http.MethodGet, "/token", nil)
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(tokenHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, generatedTestToken, recorder.Body.String())
	assert.Equal(t, http.StatusOK, recorder.Code)
}
