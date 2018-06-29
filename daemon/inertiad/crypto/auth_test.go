package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPIPrivateKey(t *testing.T) {
	key, err := getAPIPrivateKeyFromPath(nil, TestInertiaKeyPath)
	assert.Nil(t, err)
	assert.Contains(t, string(key.([]byte)), "user: git, name: ssh-public-keys")
}

func TestGetGithubKey(t *testing.T) {
	pemFile, err := os.Open(TestInertiaKeyPath)
	assert.Nil(t, err)
	_, err = GetGithubKey(pemFile)
	assert.Nil(t, err)
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(TestPrivateKey)
	assert.Nil(t, err, "generateToken must not fail")
	assert.Equal(t, token, TestToken)

	otherToken, err := GenerateToken([]byte("another_sekrit_key"))
	assert.Nil(t, err)
	assert.NotEqual(t, token, otherToken)
}
