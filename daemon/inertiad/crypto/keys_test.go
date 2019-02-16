package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPIPrivateKey(t *testing.T) {
	key, err := getAPIPrivateKeyFromPath(nil, TestInertiaKeyPath)
	assert.NoError(t, err)
	assert.Contains(t, string(key.([]byte)), "user: git, name: ssh-public-keys")
}

func TestGetGithubKey(t *testing.T) {
	pemFile, err := os.Open(TestInertiaKeyPath)
	assert.NoError(t, err)
	_, err = GetGithubKey(pemFile)
	assert.NoError(t, err)
}
