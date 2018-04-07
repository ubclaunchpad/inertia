package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
	git "gopkg.in/src-d/go-git.v4"
)

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
