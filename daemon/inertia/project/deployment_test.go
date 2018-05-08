package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
	git "gopkg.in/src-d/go-git.v4"
)

func TestSetConfig(t *testing.T) {
	deployment := &Deployment{}
	deployment.SetConfig(DeploymentConfig{
		ProjectName: "wow",
		Branch:      "amazing",
		BuildType:   "best",
	})

	assert.Equal(t, "wow", deployment.project)
	assert.Equal(t, "amazing", deployment.branch)
	assert.Equal(t, "best", deployment.buildType)
}

func TestGetBranch(t *testing.T) {
	deployment := &Deployment{branch: "master"}
	assert.Equal(t, "master", deployment.GetBranch())
}

func TestCompareRemotes(t *testing.T) {
	// Traverse back down to root directory of repository
	repo, err := git.PlainOpen("../../../")
	assert.Nil(t, err)

	deployment := &Deployment{repo: repo}

	for _, url := range urlVariations {
		err = deployment.CompareRemotes(url)
		assert.Nil(t, err)
	}
}
