package common

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testPrivateKey = []byte("very_sekrit_key")
	testToken      = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.AqFWnFeY9B8jj7-l3z0a9iaZdwIca7xhUF3fuaJjU90"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(testPrivateKey)
	assert.Nil(t, err, "generateToken must not fail")
	assert.Equal(t, token, testToken)
}

func TestGetSSHRemoteURL(t *testing.T) {
	httpsURL := "https://github.com/ubclaunchpad/inertia.git"
	sshURL := "git@github.com:ubclaunchpad/inertia.git"

	assert.Equal(t, sshURL, GetSSHRemoteURL(httpsURL))
	assert.Equal(t, sshURL, GetSSHRemoteURL(sshURL))
}

func TestCheckForGit(t *testing.T) {
	cwd, _ := os.Getwd()
	assert.NotEqual(t, nil, CheckForGit(cwd))
	inertia := strings.TrimSuffix(cwd, "/common")
	assert.Equal(t, nil, CheckForGit(inertia))
}

func TestCheckForDockerCompose(t *testing.T) {
	cwd, _ := os.Getwd()
	assert.NotEqual(t, nil, CheckForDockerCompose(cwd))
	file, _ := os.Create(cwd + "/docker-compose.yml")
	file.Close()
	assert.Equal(t, nil, CheckForDockerCompose(cwd))
	os.Remove(cwd + "/docker-compose.yml")
	file, _ = os.Create(cwd + "/docker-compose.yaml")
	file.Close()
	assert.Equal(t, nil, CheckForDockerCompose(cwd))
	os.Remove(cwd + "/docker-compose.yaml")
}
