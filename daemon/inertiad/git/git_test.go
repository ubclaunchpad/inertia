package git

import (
	"os"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

const (
	inertiaDeployTest = "https://github.com/ubclaunchpad/inertia-deploy-test.git"
)

func TestCloneIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var dir = "./test_clone/"
	repo, err := clone(inertiaDeployTest, RepoOptions{
		Directory: dir,
		Branch:    "dev",
	}, os.Stdout)
	defer os.RemoveAll(dir)
	assert.NoError(t, err)

	head, err := repo.Head()
	assert.NoError(t, err)
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
	assert.NoError(t, err)

	// Try switching branches
	err = UpdateRepository(repo, RepoOptions{Branch: "master"}, os.Stdout)
	assert.NoError(t, err)
	err = UpdateRepository(repo, RepoOptions{Branch: "dev"}, os.Stdout)
	assert.NoError(t, err)
}
