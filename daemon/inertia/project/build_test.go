package project

import (
	"context"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

// killTestContainers is a helper for tests - it implements project.ContainerStopper
func killTestContainers(cli *docker.Client, w io.Writer) error {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	// Take down all containers except the testvps
	for _, container := range containers {
		if container.Names[0] != "/testvps" {
			err := cli.ContainerKill(ctx, container.ID, "SIGKILL")
			if err != nil {
				return err
			}
		}
	}

	// Prune images
	_, err = cli.ContainersPrune(ctx, filters.Args{})
	return err
}

func TestDockerComposeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	// Set up
	err = killTestContainers(cli, nil)
	assert.Nil(t, err)

	testProjectDir := path.Join(
		os.Getenv("GOPATH"),
		"/src/github.com/ubclaunchpad/inertia/test/build/docker-compose",
	)
	testProjectName := "test_dockercompose"
	d := &Deployment{
		directory: testProjectDir,
		project:   testProjectName,
		buildType: "docker-compose",

		builders: map[string]Builder{
			"herokuish":      herokuishBuild,
			"dockerfile":     dockerBuild,
			"docker-compose": dockerCompose,
		},
		containerStopper: killTestContainers,
	}

	// Execute build
	deploy, err := dockerCompose(d, cli, os.Stdout)
	assert.Nil(t, err)

	// Execute deploy
	err = deploy()
	assert.Nil(t, err)

	// Arbitrary wait for containers to start
	time.Sleep(5 * time.Second)

	// Check for containers
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

	// try again if project no up (workaround for Travis)
	if !foundP {
		time.Sleep(20 * time.Second)
		containers, err = cli.ContainerList(
			context.Background(),
			types.ContainerListOptions{},
		)
		assert.Nil(t, err)
		for _, c := range containers {
			if strings.Contains(c.Names[0], testProjectName) {
				foundP = true
			}
		}
	}

	assert.True(t, foundDC, "docker-compose container should be active")
	assert.True(t, foundP, "project container should be active")

	// Attempt another deploy using Deployment
	err = d.Deploy(cli, os.Stdout, DeployOptions{SkipUpdate: true})
	assert.Nil(t, err)

	// Arbitrary wait for containers to start again
	time.Sleep(5 * time.Second)

	// Check for containers
	containers, err = cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	assert.Nil(t, err)
	foundDC = false
	foundP = false
	for _, c := range containers {
		if strings.Contains(c.Names[0], "docker-compose") {
			foundDC = true
		}
		if strings.Contains(c.Names[0], testProjectName) {
			foundP = true
		}
	}

	// try again if project no up (workaround for Travis)
	if !foundP {
		time.Sleep(20 * time.Second)
		containers, err = cli.ContainerList(
			context.Background(),
			types.ContainerListOptions{},
		)
		assert.Nil(t, err)
		for _, c := range containers {
			if strings.Contains(c.Names[0], testProjectName) {
				foundP = true
			}
		}
	}

	err = killTestContainers(cli, nil)
	assert.Nil(t, err)
}

func TestDockerBuildIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	err = killTestContainers(cli, nil)
	assert.Nil(t, err)

	testProjectDir := path.Join(
		os.Getenv("GOPATH"),
		"/src/github.com/ubclaunchpad/inertia/test/build/dockerfile",
	)
	testProjectName := "test_dockerfile"
	d := &Deployment{
		directory: testProjectDir,
		project:   testProjectName,
		buildType: "dockerfile",

		builders: map[string]Builder{
			"herokuish":      herokuishBuild,
			"dockerfile":     dockerBuild,
			"docker-compose": dockerCompose,
		},
		containerStopper: killTestContainers,
	}

	// Execute build
	deploy, err := dockerBuild(d, cli, os.Stdout)
	assert.Nil(t, err)

	// Execute deploy
	err = deploy()
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

	// Attempt another deploy using Deployment
	err = d.Deploy(cli, os.Stdout, DeployOptions{SkipUpdate: true})
	assert.Nil(t, err)

	// Arbitrary wait for containers to start
	time.Sleep(5 * time.Second)

	containers, err = cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	assert.Nil(t, err)
	foundP = false
	for _, c := range containers {
		if strings.Contains(c.Names[0], testProjectName) {
			foundP = true
		}
	}
	assert.True(t, foundP, "project container should be active")

	err = killTestContainers(cli, nil)
	assert.Nil(t, err)
}

func TestHerokuishBuildIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	err = killTestContainers(cli, nil)
	assert.Nil(t, err)

	testProjectDir := path.Join(
		os.Getenv("GOPATH"),
		"/src/github.com/ubclaunchpad/inertia/test/build/herokuish",
	)
	testProjectName := "test_herokuish"
	d := &Deployment{
		directory: testProjectDir,
		project:   testProjectName,
		buildType: "herokuish",

		builders: map[string]Builder{
			"herokuish":      herokuishBuild,
			"dockerfile":     dockerBuild,
			"docker-compose": dockerCompose,
		},
		containerStopper: killTestContainers,
	}

	// Execute build
	deploy, err := herokuishBuild(d, cli, os.Stdout)
	assert.Nil(t, err)

	// Execute deploy
	err = deploy()
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

	// Attempt another deploy using Deployment
	err = d.Deploy(cli, os.Stdout, DeployOptions{SkipUpdate: true})
	assert.Nil(t, err)

	// Arbitrary wait for containers to start
	time.Sleep(5 * time.Second)

	containers, err = cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	assert.Nil(t, err)
	foundP = false
	for _, c := range containers {
		if strings.Contains(c.Names[0], testProjectName) {
			foundP = true
		}
	}
	assert.True(t, foundP, "project container should be active")

	err = killTestContainers(cli, nil)
	assert.Nil(t, err)
}
