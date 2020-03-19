package cfg

import "os"

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
	return &Config{
		SecretsDirectory:     os.Getenv("INERTIA_SECRETS_DIR"),
		DataDirectory:        os.Getenv("INERTIA_DATA_DIR"),
		DockerComposeVersion: os.Getenv("INERTIA_DOCKERCOMPOSE"),
		ProjectDirectory:     os.Getenv("INERTIA_PROJECT_DIR"),
		PersistDirectory:     os.Getenv("INERTIA_PERSIST_DIR"),
	}
}
