package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/common"
)

var (
	testPrivateKey = []byte("very_sekrit_key")
	testToken      = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.AqFWnFeY9B8jj7-l3z0a9iaZdwIca7xhUF3fuaJjU90"
)

func getFakeAPIKey(tok *jwt.Token) (interface{}, error) {
	return testPrivateKey, nil
}
func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(testPrivateKey)
	assert.Nil(t, err, "generateToken must not fail")
	assert.Equal(t, token, testToken)
}

func TestAuthorizationOK(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health-check", nil)

	// Set the token for authorization.
	bearerTokenString := fmt.Sprintf("Bearer %s", testToken)
	req.Header.Set("Authorization", bearerTokenString)
	rr := httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler := http.HandlerFunc(Authorized(HealthCheckHandler, getFakeAPIKey))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, rr.Body.String(), common.DaemonOkResp)
}

func TestAuthorizationMalformedBearerString(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health-check", nil)

	// Set the token for authorization.
	req.Header.Set("Authorization", "Beare")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Authorized(HealthCheckHandler, getFakeAPIKey))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusForbidden)
	assert.Equal(t, rr.Body.String(), malformedAuthStringErrorMsg+"\n")
}

func TestAuthorizationTooManySegments(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health-check", nil)

	// Set the token for authorization.
	req.Header.Set("Authorization", "Bearer a.b.c.d")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Authorized(HealthCheckHandler, getFakeAPIKey))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusForbidden)
}

func TestAuthorizationSignatureInvalid(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health-check", nil)

	// Break the last component of the token (signature).
	bearerTokenString := fmt.Sprintf("Bearer %s", testToken+"0")
	req.Header.Set("Authorization", bearerTokenString)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Authorized(HealthCheckHandler, getFakeAPIKey))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusForbidden)
}

func TestGetGithubKey(t *testing.T) {
	inertiaKeyPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/ubclaunchpad/inertia/test_env/test_key")
	pemFile, err := os.Open(inertiaKeyPath)
	assert.Nil(t, err)
	_, err = GetGithubKey(pemFile)
	assert.Nil(t, err)
}
