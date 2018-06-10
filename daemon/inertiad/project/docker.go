package project

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
)

var (
	// ErrNoContainers is the response to indicate that no containers are active
	ErrNoContainers = errors.New("There are currently no active containers")
)

// LogOptions is used to configure retrieved container logs
type LogOptions struct {
	Container    string
	Stream       bool
	Detailed     bool
	NoTimestamps bool
}

// ContainerLogs get logs ;)
func ContainerLogs(cli *docker.Client, opts LogOptions) (io.ReadCloser, error) {
	ctx := context.Background()
	return cli.ContainerLogs(ctx, opts.Container, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     opts.Stream,
		Timestamps: !opts.NoTimestamps,
		Details:    opts.Detailed,
	})
}

// getActiveContainers returns all active containers and returns and error
// if the Daemon is the only active container
func getActiveContainers(cli *docker.Client) ([]types.Container, error) {
	containers, err := cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	if err != nil {
		return nil, err
	}

	// Error if only daemon is active
	if len(containers) == 0 || (len(containers) == 1 &&
		strings.Contains(containers[0].Names[0], "intertia-daemon")) {
		return nil, ErrNoContainers
	}

	return containers, nil
}

type containerStopper func(*docker.Client, io.Writer) error

// stopActiveContainers kills all active project containers (ie not including daemon)
func stopActiveContainers(cli *docker.Client, out io.Writer) error {
	fmt.Fprintln(out, "Shutting down active containers...")
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	// Gracefully take down all containers except the daemon
	for _, container := range containers {
		if container.Names[0] != "/inertia-daemon" {
			fmt.Fprintln(out, "Stopping "+container.Names[0]+"...")
			timeout := 10 * time.Second
			err := cli.ContainerStop(ctx, container.ID, &timeout)
			if err != nil {
				return err
			}
		}
	}

	// Prune images
	_, err = cli.ContainersPrune(ctx, filters.Args{})
	return err
}

// StreamContainerLogs streams logs from given container ID. Best used as a
// goroutine.
func StreamContainerLogs(client *docker.Client, id string, out io.Writer,
	stop chan struct{}) error {
	// Attach logs and report build progress until container exits
	reader, err := ContainerLogs(client, LogOptions{
		Container: id, Stream: true,
		NoTimestamps: true,
	})
	if err != nil {
		return err
	}
	defer reader.Close()
	log.FlushRoutine(out, reader, stop)
	return nil
}
