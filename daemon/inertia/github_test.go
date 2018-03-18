package main

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGithubKey(t *testing.T) {
	inertiaKeyPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/ubclaunchpad/inertia/test_env/test_key")
	pemFile, err := os.Open(inertiaKeyPath)
	assert.Nil(t, err)
	_, err = getGithubKey(pemFile)
	assert.Nil(t, err)
}
