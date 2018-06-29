package project

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/build"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	git "gopkg.in/src-d/go-git.v4"

	docker "github.com/docker/docker/client"
)

type MockBuilder struct {
	builder func() error
}

func (m *MockBuilder) Build(string, *build.Config, *docker.Client, io.Writer) (func() error, error) {
	return m.builder, nil
}

func (m *MockBuilder) GetBuildStageName() string { return "build" }

func TestSetConfig(t *testing.T) {
	deployment := &Deployment{}
	deployment.SetConfig(DeploymentConfig{
		ProjectName:   "wow",
		Branch:        "amazing",
		BuildType:     "best",
		BuildFilePath: "/robertcompose.yml",
	})

	assert.Equal(t, "wow", deployment.project)
	assert.Equal(t, "amazing", deployment.branch)
	assert.Equal(t, "best", deployment.buildType)
	assert.Equal(t, "/robertcompose.yml", deployment.buildFilePath)
}

func TestDeployMockSkipUpdateIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	buildCalled := false
	stopCalled := false
	d := Deployment{
		directory: "./test/",
		buildType: "test",
		builder: &MockBuilder{
			builder: func() error {
				buildCalled = true
				return nil
			},
		},
		containerStopper: func(*docker.Client, io.Writer) error {
			stopCalled = true
			return nil
		},
	}

	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	err = d.Deploy(cli, os.Stdout, DeployOptions{SkipUpdate: true})
	assert.Nil(t, err)
	assert.Equal(t, true, buildCalled)
	assert.Equal(t, true, stopCalled)
}

func TestDownIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	called := false
	d := Deployment{
		directory: "./test/",
		buildType: "test",
		containerStopper: func(*docker.Client, io.Writer) error {
			called = true
			return nil
		},
	}

	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	err = d.Down(cli, os.Stdout)
	if err != containers.ErrNoContainers {
		assert.Nil(t, err)
	}

	assert.True(t, called)
}

func TestGetStatusIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Traverse back down to root directory of repository
	repo, err := git.PlainOpen("../../../")
	assert.Nil(t, err)

	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	deployment := &Deployment{
		repo:      repo,
		buildType: "test",
		builder:   &MockBuilder{},
	}
	status, err := deployment.GetStatus(cli)
	assert.Nil(t, err)
	assert.False(t, status.BuildContainerActive)
	assert.Equal(t, "test", status.BuildType)
}

func TestGetBranch(t *testing.T) {
	deployment := &Deployment{branch: "master"}
	assert.Equal(t, "master", deployment.GetBranch())
}

func TestCompareRemotes(t *testing.T) {
	urlVariations := []string{
		"https://github.com/ubclaunchpad/inertia.git",
		"git://github.com/ubclaunchpad/inertia.git",
	}

	// Traverse back down to root directory of repository
	repo, err := git.PlainOpen("../../../")
	assert.Nil(t, err)

	deployment := &Deployment{repo: repo}
	for _, url := range urlVariations {
		err = deployment.CompareRemotes(url)
		assert.Nil(t, err)
	}
}
