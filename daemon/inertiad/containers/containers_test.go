package containers

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

func TestContainerLogs(t *testing.T) {
	cli, err := NewDockerClient()
	assert.Nil(t, err)
	defer cli.Close()

	type args struct {
		opts LogOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"successfully get logs", args{
			LogOptions{Container: "/testvps"}}, false},
		{"successfully get logs with lines", args{
			LogOptions{Container: "/testvps", Entries: 100}}, false},
		{"successfully get logs without leading slash", args{
			LogOptions{Container: "testvps", Entries: 100}}, false},
		{"fail on unknown container", args{
			LogOptions{Container: "asdf"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ContainerLogs(cli, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestStreamContainerLogs(t *testing.T) {
	cli, err := NewDockerClient()
	assert.Nil(t, err)
	defer cli.Close()

	// todo: flesh this out a bit more
	stop := make(chan struct{})
	go StreamContainerLogs(cli, "/testvps", os.Stdout, stop)
	time.Sleep(1 * time.Second)
	close(stop)
}

func TestGetActiveContainers(t *testing.T) {
	cli, err := NewDockerClient()
	assert.Nil(t, err)
	defer cli.Close()

	_, err = GetActiveContainers(cli)
	assert.Nil(t, err)
}

func TestPrune(t *testing.T) {
	cli, err := NewDockerClient()
	assert.Nil(t, err)
	defer cli.Close()

	Prune(cli)
}

func TestPruneAll(t *testing.T) {
	cli, err := NewDockerClient()
	assert.Nil(t, err)
	defer cli.Close()

	PruneAll(cli, "docker/compose")

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
