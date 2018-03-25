package auth

import (
	"fmt"
	"net/http"
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

func testHealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, common.DaemonOkResp)
}

func getFakeAPIKey(tok *jwt.Token) (interface{}, error) {
	return testPrivateKey, nil
}
func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(testPrivateKey)
	assert.Nil(t, err, "generateToken must not fail")
	assert.Equal(t, token, testToken)
}

func TestGetGithubKey(t *testing.T) {
	inertiaKeyPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/ubclaunchpad/inertia/test_env/test_key")
	pemFile, err := os.Open(inertiaKeyPath)
	assert.Nil(t, err)
	_, err = GetGithubKey(pemFile)
	assert.Nil(t, err)
}
