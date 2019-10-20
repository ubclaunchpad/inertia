package daemon

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/api"
)

func TestTokenHandler(t *testing.T) {
	var (
		generatedTestToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzZXNzaW9uX2lkIjoiIiwidXNlciI6Im1hc3RlciIsImFkbWluIjp0cnVlLCJleHBpcnkiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiJ9.BtWKKZsdqGXDoTZPV6id5PkS0VEVUaLzIangLr2c2XM"
		testInertiaKeyPath = "../../../test/keys/id_rsa"
	)
	os.Setenv("INERTIA_GH_KEY_PATH", testInertiaKeyPath)
	defer os.Setenv("INERTIA_GH_KEY_PATH", "")

	// Assemble request
	req, err := http.NewRequest(http.MethodGet, "/token", nil)
	assert.NoError(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(tokenHandler)
	handler.ServeHTTP(recorder, req)

	var token string
	api.Unmarshal(recorder.Body, api.KV{Key: "token", Value: &token})
	assert.Equal(t, generatedTestToken, token)
	assert.Equal(t, http.StatusOK, recorder.Code)
}
