package common

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomString(t *testing.T) {
	randString, err := GenerateRandomString()
	assert.Nil(t, err)
	assert.Equal(t, len(randString), 44)
}

func TestCheckForDockerCompose(t *testing.T) {
	cwd, err := os.Getwd()
	assert.Nil(t, err)

	ymlPath := path.Join(cwd, "/docker-compose.yml")
	yamlPath := path.Join(cwd, "/docker-compose.yaml")

	// No!
	b := CheckForDockerCompose(cwd)
	assert.False(t, b)

	// Yes!
	file, err := os.Create(ymlPath)
	assert.Nil(t, err)
	file.Close()
	b = CheckForDockerCompose(cwd)
	assert.True(t, b)
	os.Remove(ymlPath)

	// Yes!
	file, err = os.Create(yamlPath)
	assert.Nil(t, err)
	file.Close()
	b = CheckForDockerCompose(cwd)
	assert.True(t, b)
	os.Remove(yamlPath)
}

func TestCheckForDockerfile(t *testing.T) {
	cwd, err := os.Getwd()
	assert.Nil(t, err)

	path := path.Join(cwd, "/Dockerfile")

	// No!
	b := CheckForDockerfile(cwd)
	assert.False(t, b)

	// Yes!
	file, err := os.Create(path)
	assert.Nil(t, err)
	file.Close()
	b = CheckForDockerfile(cwd)
	assert.True(t, b)
	os.Remove(path)
}

func TestExtract(t *testing.T) {
	for _, url := range remoteURLVariations {
		repoName, err := ExtractRepository(url)
		assert.Nil(t, err)
		assert.Equal(t, "ubclaunchpad/inertia", repoName)
	}

	repoNameWithHyphens, err := ExtractRepository("git@github.com:ubclaunchpad/inertia-deploy-test.git")
	assert.Nil(t, err)
	assert.Equal(t, "ubclaunchpad/inertia-deploy-test", repoNameWithHyphens)

	repoNameWithDots, err := ExtractRepository("git@github.com:ubclaunchpad/inertia.deploy.test.git")
	assert.Nil(t, err)
	assert.Equal(t, "ubclaunchpad/inertia.deploy.test", repoNameWithDots)

	repoNameWithMixed, err := ExtractRepository("git@github.com:ubclaunchpad/inertia-deploy.test.git")
	assert.Nil(t, err)
	assert.Equal(t, "ubclaunchpad/inertia-deploy.test", repoNameWithMixed)

	repoNameWithErr, err := ExtractRepository("asdfasda")
	assert.NotNil(t, err)
	assert.Equal(t, "$YOUR_REPOSITORY", repoNameWithErr)
}
