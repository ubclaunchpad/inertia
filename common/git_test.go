package common

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var urlVariations = []string{
	"git@github.com:ubclaunchpad/inertia.git",
	"https://github.com/ubclaunchpad/inertia.git",
	"git://github.com/ubclaunchpad/inertia.git",
	"git://github.com/ubclaunchpad/inertia",
}

func TestCheckForGit(t *testing.T) {
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	assert.NotEqual(t, nil, CheckForGit(cwd))
	inertia, _ := path.Split(cwd)
	assert.Equal(t, nil, CheckForGit(inertia))
}

func TestGetSSHRemoteURL(t *testing.T) {
	for _, url := range urlVariations {
		assert.Equal(t, urlVariations[0], GetSSHRemoteURL(url))
	}
}

func TestGetBranchFromRef(t *testing.T) {
	branch := GetBranchFromRef("refs/heads/master")
	assert.Equal(t, "master", branch)
}
