package git

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRepoRemote(t *testing.T) {
	url, err := GetRepoRemote("origin")
	assert.NoError(t, err)
	assert.Contains(t, url, "ubclaunchpad/inertia")
}

func TestGetRepoCurrentBranch(t *testing.T) {
	// This test does not work on Travis, since Travis cloning doesn't always
	// set up branches correctly (typically detached) - same with Appveyor
	if os.Getenv("CI") == "true" || os.Getenv("TRAVIS") == "true" || os.Getenv("APPVEYOR") == "True" {
		t.Skip("skipping test because of CI")
	}
	_, err := GetRepoCurrentBranch()
	assert.NoError(t, err)
}

func TestIsRepo(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	assert.Error(t, IsRepo(cwd))
	assert.NoError(t, IsRepo("../../"))
}
