package containers

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestContainerLogs(t *testing.T) {
	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	_, err = ContainerLogs(cli, LogOptions{Container: "/testvps"})
	assert.Nil(t, err)
}

func TestStreamContainerLogs(t *testing.T) {
	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	// todo: flesh this out a bit more
	stop := make(chan struct{})
	go StreamContainerLogs(cli, "/testvps", os.Stdout, stop)
	time.Sleep(1 * time.Second)
	close(stop)
}

func TestGetActiveContainers(t *testing.T) {
	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	_, err = GetActiveContainers(cli)
	assert.Nil(t, err)
}

func TestPrune(t *testing.T) {
	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	err = Prune(cli)
	assert.Nil(t, err)
}

func TestPruneAll(t *testing.T) {
	cli, err := docker.NewEnvClient()
	assert.Nil(t, err)
	defer cli.Close()

	err = PruneAll(cli, "gliderlabs/herokuish", "docker/compose")
	assert.Nil(t, err)

	// Exceptions should still be present
	found := false
	list, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	assert.Nil(t, err)
	for _, i := range list {
		if strings.Contains(i.RepoTags[0], "docker/compose") {
			found = true
		}
	}
	assert.True(t, found)
}
