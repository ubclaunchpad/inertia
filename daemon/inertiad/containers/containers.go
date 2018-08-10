package containers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
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
	Entries      int
}

// ContainerLogs get logs ;)
func ContainerLogs(docker *docker.Client, opts LogOptions) (io.ReadCloser, error) {
	ctx := context.Background()
	return docker.ContainerLogs(ctx, opts.Container, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     opts.Stream,
		Timestamps: !opts.NoTimestamps,
		Details:    opts.Detailed,
		Tail:       strconv.Itoa(opts.Entries),
	})
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

// GetActiveContainers returns all active containers and returns and error
// if the Daemon is the only active container
func GetActiveContainers(docker *docker.Client) ([]types.Container, error) {
	containers, err := docker.ContainerList(
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

// ContainerStopper is a function interface
type ContainerStopper func(*docker.Client, io.Writer) error

// StopActiveContainers kills all active project containers (ie not including daemon)
func StopActiveContainers(docker *docker.Client, out io.Writer) error {
	fmt.Fprintln(out, "Shutting down active containers...")
	ctx := context.Background()
	containers, err := docker.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	// Gracefully take down all containers except the daemon
	for _, container := range containers {
		if container.Names[0] != "/inertia-daemon" {
			fmt.Fprintln(out, "Stopping "+container.Names[0]+"...")
			timeout := 10 * time.Second
			if err := docker.ContainerStop(ctx, container.ID, &timeout); err != nil {
				return err
			}

			// Archive container
			docker.ContainerRename(
				ctx, container.ID, fmt.Sprintf("%s-%d", container.Names[0], time.Now().Unix()))
		}
	}
	return nil
}

// Prune clears up unused Docker assets.
func Prune(docker *docker.Client) error {
	ctx := context.Background()

	_, errImages := docker.ImagesPrune(ctx, filters.Args{})
	_, errContainers := docker.ContainersPrune(ctx, filters.Args{})
	_, errVolumes := docker.VolumesPrune(ctx, filters.Args{})
	if errImages != nil || errContainers != nil || errVolumes != nil {
		return fmt.Errorf(
			"Errors encountered: %s ; %s ; %s",
			errImages, errContainers, errVolumes,
		)
	}
	return nil
}

// PruneAll forcibly removes all images except given exceptions (repo tag names)
func PruneAll(docker *docker.Client, exceptions ...string) error {
	args := filters.NewArgs()
	ctx := context.Background()

	// Delete images
	list, err := docker.ImageList(ctx, types.ImageListOptions{
		Filters: args,
		All:     true,
	})
	if err != nil {
		return err
	}
	for _, i := range list {
		delete := true
		for _, e := range exceptions {
			if strings.Contains(i.RepoTags[0], e) {
				delete = false
			}
		}
		if delete {
			docker.ImageRemove(ctx, i.ID, types.ImageRemoveOptions{
				Force:         true,
				PruneChildren: true,
			})
		}
	}

	// Perform basic prune on containers and volumes
	docker.ContainersPrune(ctx, filters.Args{})
	docker.VolumesPrune(ctx, filters.Args{})
	return nil
}

// Wait blocks until given container ID stops
func Wait(cli *docker.Client, id string, stop chan struct{}) (int64, error) {
	var status container.ContainerWaitOKBody
	statusCh, errCh := cli.ContainerWait(context.Background(), id, "")
	select {
	case err := <-errCh:
		if err != nil {
			return 0, err
		}
	case status = <-statusCh:
		// Exit log stream
		close(stop)
	}
	return status.StatusCode, nil
}

// StartAndWait starts and waits for container to exit
func StartAndWait(cli *docker.Client, containerID string, out io.Writer) error {
	ctx := context.Background()
	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	stop := make(chan struct{})
	go StreamContainerLogs(cli, containerID, out, stop)
	exitCode, err := Wait(cli, containerID, stop)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("Container exited with non-zero status %d", exitCode)
	}

	return nil
}
