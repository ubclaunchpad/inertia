package project

import (
	"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestDockerComposeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	testProjectDir := path.Join(
		os.Getenv("GOPATH"),
		"/src/github.com/ubclaunchpad/inertia/test/build/docker-compose",
	)
	testProjectName := "test_dockercompose"
	d := &Deployment{
		directory: testProjectDir,
		project:   testProjectName,
		buildType: "docker-compose",
	}
	d.Down(cli, os.Stdout)

	// Execute build
	err = dockerCompose(d, cli, os.Stdout)
	assert.Nil(t, err)

	// Arbitrary wait for containers to start
	time.Sleep(5 * time.Second)

	containers, err := cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	assert.Nil(t, err)
	foundDC := false
	foundP := false
	for _, c := range containers {
		if strings.Contains(c.Names[0], "docker-compose") {
			foundDC = true
		}
		if strings.Contains(c.Names[0], testProjectName) {
			foundP = true
		}
	}
	assert.True(t, foundDC, "docker-compose container should be active")
	assert.True(t, foundP, "project container should be active")

	err = d.Down(cli, os.Stdout)
	assert.Nil(t, err)
}

func TestDockerBuildIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	testProjectDir := path.Join(
		os.Getenv("GOPATH"),
		"/src/github.com/ubclaunchpad/inertia/test/build/dockerfile",
	)
	testProjectName := "test_dockerfile"
	d := &Deployment{
		directory: testProjectDir,
		project:   testProjectName,
		buildType: "dockerfile",
	}
	d.Down(cli, os.Stdout)

	// Execute build
	err = dockerBuild(d, cli, os.Stdout)
	assert.Nil(t, err)

	// Arbitrary wait for containers to start
	time.Sleep(5 * time.Second)

	containers, err := cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	assert.Nil(t, err)
	foundP := false
	for _, c := range containers {
		if strings.Contains(c.Names[0], testProjectName) {
			foundP = true
		}
	}
	assert.True(t, foundP, "project container should be active")

	err = d.Down(cli, os.Stdout)
	assert.Nil(t, err)
}

func TestHerokuishBuildIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	testProjectDir := path.Join(
		os.Getenv("GOPATH"),
		"/src/github.com/ubclaunchpad/inertia/test/build/herokuish",
	)
	testProjectName := "test_herokuish"
	d := &Deployment{
		directory: testProjectDir,
		project:   testProjectName,
		buildType: "herokuish",
	}
	d.Down(cli, os.Stdout)

	// Execute build
	err = herokuishBuild(d, cli, os.Stdout)
	assert.Nil(t, err)

	// Arbitrary wait for containers to start
	time.Sleep(5 * time.Second)

	containers, err := cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	assert.Nil(t, err)
	foundP := false
	for _, c := range containers {
		if strings.Contains(c.Names[0], testProjectName) {
			foundP = true
		}
	}
	assert.True(t, foundP, "project container should be active")

	// err = d.Down(cli, os.Stdout)
	// assert.Nil(t, err)
}
