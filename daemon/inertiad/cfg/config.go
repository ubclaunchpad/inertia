package cfg

import "os"

// Config provides basic daemon configuration
type Config struct {
	// Directories
	ProjectDirectory string // "/app/host/inertia/project"
	SSLDirectory     string // "/app/host/inertia/config/ssl/"
	DataDirectory    string // "/app/host/inertia/data/"

	// Build tools
	DockerComposeVersion string // "docker/compose:1.21.0"
	HerokuishVersion     string // "gliderlabs/herokuish:v0.4.0"
}

// New creates a new daemon configuration from environment values
func New() *Config {
	return &Config{
		SSLDirectory:         os.Getenv("INERTIA_SSL_DIR"),
		DataDirectory:        os.Getenv("INERTIA_DATA_DIR"),
		DockerComposeVersion: os.Getenv("INERTIA_DOCKERCOMPOSE"),
		HerokuishVersion:     os.Getenv("INERTIA_HEROKUISH"),
		ProjectDirectory:     os.Getenv("INERTIA_PROJECT_DIR"),
	}
}
