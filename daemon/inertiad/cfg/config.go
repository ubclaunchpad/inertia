package cfg

import (
	"context"
	"fmt"
	"os"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
)

// Config provides basic daemon configuration
type Config struct {
	// Directories
	ProjectDirectory string // "/app/host/inertia/project/"
	PersistDirectory string // "/app/host/inertia/persist"
	DataDirectory    string // "/app/host/inertia/data/"
	SecretsDirectory string // "/app/host/.inertia/"

	// Build tools
	DockerComposeVersion string // "docker/compose:${version}"

	WebhookSecret string
}

// New creates a new daemon configuration from environment values
func New() *Config {
	dcVersionString := "latest"
	dcVersion, err := containers.GetLatestImageTag(context.TODO(), "docker/compose", nil)
	if err == nil {
		dcVersionString = dcVersion.String()
	}

	return &Config{
		SecretsDirectory:     os.Getenv("INERTIA_SECRETS_DIR"),
		DataDirectory:        os.Getenv("INERTIA_DATA_DIR"),
		DockerComposeVersion: fmt.Sprintf("docker/compose:%s", dcVersionString),
		ProjectDirectory:     os.Getenv("INERTIA_PROJECT_DIR"),
		PersistDirectory:     os.Getenv("INERTIA_PERSIST_DIR"),
	}
}
