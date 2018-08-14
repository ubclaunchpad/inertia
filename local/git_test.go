package local

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRepoRemote(t *testing.T) {
	url, err := GetRepoRemote("origin")
	assert.Nil(t, err)
	assert.Contains(t, url, "ubclaunchpad/inertia")
}

func TestGetRepoCurrentBranch(t *testing.T) {
	// This test does not work on Travis, since Travis cloning doesn't always
	// set up branches correctly (typically detached) - same with Appveyor
	if os.Getenv("TRAVIS") == "true" || os.Getenv("APPVEYOR") == "True" {
		t.Skip("skipping test because of CI")
	}
	_, err := GetRepoCurrentBranch()
	if err != nil {
		fmt.Print(err)
	}
	assert.Nil(t, err)
}

func TestCheckForGit(t *testing.T) {
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	assert.NotNil(t, checkForGit(cwd))
	inertia, _ := filepath.Split(cwd)
	assert.Nil(t, checkForGit(inertia))
}
