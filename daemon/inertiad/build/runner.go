package build

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
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
	return containers.WatchContainers(client, nil)
}
