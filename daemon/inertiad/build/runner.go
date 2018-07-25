package build

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
)

// run starts project and tracks all active project containers and pipes an error
// to the returned channel if any container exits or errors.
func run(ctx context.Context, client *docker.Client, id string, out io.Writer) <-chan error {
	fmt.Fprintln(out, "Starting up project...")
	if err := client.ContainerStart(ctx, id, types.ContainerStartOptions{}); err != nil {
		errCh := make(chan error, 1)
		errCh <- err
		return errCh
	}
	return watchContainers(client, nil)
}

func watchContainers(client *docker.Client, stop chan struct{}) <-chan error {
	exitCh := make(chan error, 1)
	list, err := containers.GetActiveContainers(client)
	if err != nil {
		exitCh <- err
		return exitCh
	}
	for _, c := range list {
		statusCh, errCh := client.ContainerWait(context.Background(), c.ID, "")
		go func(id string) {
			select {
			case err := <-errCh:
				if err != nil {
					exitCh <- err
					return
				}
			case status := <-statusCh:
				if status.Error != nil {
					exitCh <- fmt.Errorf(
						"container %s exited with status %d: %s", id, status.StatusCode, status.Error.Message)
				}
				exitCh <- fmt.Errorf("container %s exited with status %d", id, status.StatusCode)
				return
			}
		}(c.ID)
	}
	return exitCh
}
