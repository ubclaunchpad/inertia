package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogHandlerNoDeployment(t *testing.T) {
	// Assmble request
	req, err := http.NewRequest("POST", "/down", nil)
	assert.Nil(t, err)

	// Record responses
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(downHandler)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusPreconditionFailed)
	assert.Contains(t, recorder.Body.String(), msgNoDeployment)
}
