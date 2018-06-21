package local

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRepoRemote(t *testing.T) {
	url, err := GetRepoRemote("origin")
	assert.Nil(t, err)
	assert.Contains(t, url, "ubclaunchpad/inertia")
}

func TestGetRepoCurrentBranch(t *testing.T) {
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
	inertia, _ := path.Split(cwd)
	assert.Nil(t, checkForGit(inertia))
}
