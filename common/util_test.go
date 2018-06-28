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

func TestCheckForProcfile(t *testing.T) {
	cwd, err := os.Getwd()
	assert.Nil(t, err)

	path := path.Join(cwd, "/Procfile")

	// No!
	b := CheckForProcfile(cwd)
	assert.False(t, b)

	// Yes!
	file, err := os.Create(path)
	assert.Nil(t, err)
	file.Close()
	b = CheckForProcfile(cwd)
	assert.True(t, b)
	os.Remove(path)
}

func TestExtract(t *testing.T) {
	for _, url := range remoteURLVariations {
		repoName := ExtractRepository(url)
		assert.Equal(t, "ubclaunchpad/inertia", repoName)
	}

	repoNameWithHyphens := ExtractRepository("git@github.com:ubclaunchpad/inertia-deploy-test.git")
	assert.Equal(t, "ubclaunchpad/inertia-deploy-test", repoNameWithHyphens)

	repoNameWithDots := ExtractRepository("git@github.com:ubclaunchpad/inertia.deploy.test.git")
	assert.Equal(t, "ubclaunchpad/inertia.deploy.test", repoNameWithDots)

	repoNameWithMixed := ExtractRepository("git@github.com:ubclaunchpad/inertia-deploy.test.git")
	assert.Equal(t, "ubclaunchpad/inertia-deploy.test", repoNameWithMixed)
}

func TestParseDate(t *testing.T) {
	assert.NotNil(t, ParseDate("2006-01-02T15:04:05.000Z"))
}

func TestParseInt64(t *testing.T) {
	parsed, err := ParseInt64("10")
	assert.Nil(t, err)
	assert.Equal(t, int64(10), parsed)

	_, err = ParseInt64("")
	assert.NotNil(t, err)
}
