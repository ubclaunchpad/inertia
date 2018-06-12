package project

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	git "gopkg.in/src-d/go-git.v4"

	docker "github.com/docker/docker/client"
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

func TestDeployMockSkipUpdateIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	buildCalled := false
	stopCalled := false
	d := Deployment{
		directory: "./test/",
		buildType: "test",
		builders: map[string]Builder{
			"test": func(*Deployment, *docker.Client, io.Writer) (func() error, error) {
				return func() error {
					buildCalled = true
					return nil
				}, nil
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
	assert.True(t, buildCalled)
	assert.True(t, stopCalled)
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
	if err != ErrNoContainers {
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
	// Traverse back down to root directory of repository
	// repo, err := git.PlainOpen("../../../")
	// assert.Nil(t, err)

	// deployment := &Deployment{repo: repo}

	// for _, url := range urlVariations {
	// 	err = deployment.CompareRemotes(url)
	// 	assert.Nil(t, err)
	// }
}
