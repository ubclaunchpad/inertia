package git

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	git "gopkg.in/src-d/go-git.v4"
)

const (
	inertiaDeployTest = "https://github.com/ubclaunchpad/inertia-deploy-test.git"
)

func TestCloneIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dir := "./test_clone/"
	repo, err := clone(dir, inertiaDeployTest, "dev", nil, os.Stdout)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)

	head, err := repo.Head()
	assert.Nil(t, err)
	assert.Equal(t, "dev", head.Name().Short())
}

func TestUpdateRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dir := "./test_update/"
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: inertiaDeployTest,
	})
	defer os.RemoveAll(dir)
	assert.Nil(t, err)

	// Try switching branches
	err = UpdateRepository(dir, repo, "master", nil, os.Stdout)
	assert.Nil(t, err)
	err = UpdateRepository(dir, repo, "dev", nil, os.Stdout)
	assert.Nil(t, err)
}
