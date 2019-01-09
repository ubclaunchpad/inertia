package project

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/mocks"
	gogit "gopkg.in/src-d/go-git.v4"
)

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

func TestDeployMock(t *testing.T) {
	buildCalled := false
	stopCalled := false
	d := Deployment{
		directory: "./test/",
		buildType: "test",
		builder: &mocks.FakeBuilder{
			builder: func() error {
				buildCalled = true
				return nil
			},
			stopper: func() error {
				stopCalled = true
				return nil
			},
		},
	}

	cli, err := containers.NewDockerClient()
	assert.Nil(t, err)
	defer cli.Close()

	deploy, err := d.Deploy(cli, os.Stdout, DeployOptions{SkipUpdate: true})
	assert.Nil(t, err)

	deploy()
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
		builder: &MockBuilder{
			stopper: func() error {
				called = true
				return nil
			},
		},
	}

	cli, err := containers.NewDockerClient()
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
	repo, err := gogit.PlainOpen("../../../")
	assert.Nil(t, err)

	cli, err := containers.NewDockerClient()
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

func TestDeployment_CompareRemotes(t *testing.T) {
	repo, err := gogit.PlainOpen("../../../")
	assert.Nil(t, err)
	type args struct {
		remoteURL string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"blank arg", args{""}, false},
		{"http url matches", args{"https://github.com/ubclaunchpad/inertia.git"}, false},
		{"ssh url matches", args{"git://github.com/ubclaunchpad/inertia.git"}, false},
		{"invalid url does not match", args{"https://www.ubclaunchpad.com"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployment{repo: repo}
			if err := d.CompareRemotes(tt.args.remoteURL); (err != nil) != tt.wantErr {
				t.Errorf("Deployment.CompareRemotes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
