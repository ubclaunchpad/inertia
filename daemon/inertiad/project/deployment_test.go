package project

import (
	"io"
	"os"
	"testing"

	docker "github.com/docker/docker/client"
	gogit "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/build/mocks"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
)

func newDefaultFakeBuilder(builder func() error, stopper func() error) *mocks.FakeContainerBuilder {
	var fakeBuilder = &mocks.FakeContainerBuilder{
		PruneStub:    func(*docker.Client, io.Writer) error { return stopper() },
		PruneAllStub: func(*docker.Client, io.Writer) error { return stopper() },
	}
	fakeBuilder.GetBuildStageNameReturns("build")
	fakeBuilder.BuildReturns(builder, nil)
	return fakeBuilder
}

func TestSetConfig(t *testing.T) {
	deployment := &Deployment{}
	deployment.SetConfig(DeploymentConfig{
		ProjectName:          "wow",
		Branch:               "amazing",
		BuildType:            "best",
		BuildFilePath:        "/robertcompose.yml",
		SlackNotificationURL: "https://my.slack.url",
	})

	assert.Equal(t, "wow", deployment.project)
	assert.Equal(t, "amazing", deployment.branch)
	assert.Equal(t, "best", deployment.buildType)
	assert.Equal(t, "/robertcompose.yml", deployment.buildFilePath)
	assert.Len(t, deployment.notifiers, 1)
}

func TestDeployMock(t *testing.T) {
	var (
		buildCalled = false
		stopCalled  = false
	)
	var fakeBuilder = newDefaultFakeBuilder(
		func() error {
			buildCalled = true
			return nil
		},
		func() error {
			stopCalled = true
			return nil
		},
	)
	var d = Deployment{
		directory: "./test/",
		buildType: "test",
		builder:   fakeBuilder,
	}

	cli, err := containers.NewDockerClient()
	assert.NoError(t, err)
	defer cli.Close()

	deploy, err := d.Deploy(cli, os.Stdout, DeployOptions{SkipUpdate: true})
	assert.NoError(t, err)

	deploy()
	assert.Equal(t, true, buildCalled)
	assert.Equal(t, true, stopCalled)
}

func TestDownIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var called = false
	var fakeBuilder = newDefaultFakeBuilder(nil, func() error {
		called = true
		return nil
	})
	var d = Deployment{
		directory: "./test/",
		buildType: "test",
		builder:   fakeBuilder,
	}

	cli, err := containers.NewDockerClient()
	assert.NoError(t, err)
	defer cli.Close()

	err = d.Down(cli, os.Stdout)
	if err != containers.ErrNoContainers {
		assert.NoError(t, err)
	}

	assert.True(t, called)
}

func TestGetStatusIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Traverse back down to root directory of repository
	repo, err := gogit.PlainOpen("../../../")
	assert.NoError(t, err)

	cli, err := containers.NewDockerClient()
	assert.NoError(t, err)
	defer cli.Close()

	var fakeBuilder = newDefaultFakeBuilder(nil, nil)
	var deployment = &Deployment{
		repo:      repo,
		buildType: "test",
		builder:   fakeBuilder,
	}
	status, err := deployment.GetStatus(cli)

	assert.NoError(t, err)
	assert.False(t, status.BuildContainerActive)
	assert.Equal(t, "test", status.BuildType)
}

func TestGetBranch(t *testing.T) {
	deployment := &Deployment{branch: "master"}
	assert.Equal(t, "master", deployment.GetBranch())
}

func TestDeployment_CompareRemotes(t *testing.T) {
	repo, err := gogit.PlainOpen("../../../")
	assert.NoError(t, err)
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
