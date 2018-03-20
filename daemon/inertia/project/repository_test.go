package project

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	inertiaDeployTest = "https://github.com/ubclaunchpad/inertia-deploy-test.git"
)

var urlVariations = []string{
	"git@github.com:ubclaunchpad/inertia.git",
	"https://github.com/ubclaunchpad/inertia.git",
	"git://github.com/ubclaunchpad/inertia.git",
	"git://github.com/ubclaunchpad/inertia",
}

func getInertiaDeployTestKey() (ssh.AuthMethod, error) {
	pemFile, err := os.Open("../test_env/test_key")
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(pemFile)
	if err != nil {
		return nil, err
	}
	return ssh.NewPublicKeys("git", bytes, "")
}

func TestClone(t *testing.T) {
	dir := "./test_clone/"
	repo, err := Clone(dir, inertiaDeployTest, "dev", nil, os.Stdout)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)

	head, err := repo.Head()
	assert.Nil(t, err)
	assert.Equal(t, "dev", head.Name().Short())
}

func TestForcePull(t *testing.T) {
	dir := "./test_force_pull/"
	auth, err := getInertiaDeployTestKey()
	assert.Nil(t, err)
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: inertiaDeployTest,
	})
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	forcePulledRepo, err := ForcePull(dir, repo, auth, os.Stdout)
	assert.Nil(t, err)

	// Try switching branches
	err = UpdateRepository(dir, forcePulledRepo, "dev", auth, os.Stdout)
	assert.Nil(t, err)
	err = UpdateRepository(dir, forcePulledRepo, "master", auth, os.Stdout)
	assert.Nil(t, err)
}

func TestUpdateRepository(t *testing.T) {
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

func TestCompareRemotes(t *testing.T) {
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	inertia, _ := path.Split(cwd)

	repo, err := git.PlainOpen(inertia)
	assert.Nil(t, err)

	for _, url := range urlVariations {
		err = CompareRemotes(repo, url)
		assert.Nil(t, err)
	}
}
