package containers

import docker "github.com/docker/docker/client"

// MaxDockerVersion is the maximum supported API version
const MaxDockerVersion = "1.37"

// NewDockerClient creates a new Docker Client set to a predefined Docker API
func NewDockerClient() (*docker.Client, error) {
	return docker.NewClientWithOpts(docker.WithVersion(MaxDockerVersion))
}
