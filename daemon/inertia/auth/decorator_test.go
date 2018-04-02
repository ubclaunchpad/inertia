package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/common"
)

func TestAuthorizationOK(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health-check", nil)

	// Set the token for authorization.
	bearerTokenString := fmt.Sprintf("Bearer %s", testToken)
	req.Header.Set("Authorization", bearerTokenString)
	rr := httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler := http.HandlerFunc(Authorized(testHealthCheckHandler, getFakeAPIKey))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, rr.Body.String(), common.DaemonOkResp)
}

func TestAuthorizationMalformedBearerString(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health-check", nil)

	// Set the token for authorization.
	req.Header.Set("Authorization", "Beare")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Authorized(testHealthCheckHandler, getFakeAPIKey))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusForbidden)
	assert.Equal(t, rr.Body.String(), malformedAuthStringErrorMsg+"\n")
}

func TestAuthorizationTooManySegments(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health-check", nil)

	// Set the token for authorization.
	req.Header.Set("Authorization", "Bearer a.b.c.d")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Authorized(testHealthCheckHandler, getFakeAPIKey))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusForbidden)
}

func TestAuthorizationSignatureInvalid(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health-check", nil)

	// Break the last component of the token (signature).
	bearerTokenString := fmt.Sprintf("Bearer %s", testToken+"0")
	req.Header.Set("Authorization", bearerTokenString)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Authorized(testHealthCheckHandler, getFakeAPIKey))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusForbidden)
}
