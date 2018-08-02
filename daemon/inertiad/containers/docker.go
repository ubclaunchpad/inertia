package containers

import (
	"context"

	docker "github.com/docker/docker/client"
)

// NewDockerClient creates a new Docker Client from ENV values and negotiates
// the correct API version
func NewDockerClient() (*docker.Client, error) {
	c, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}
	c.NegotiateAPIVersion(context.Background())
	return c, nil
}
