package containers

import (
	"context"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestCleanup(t *testing.T) {
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
