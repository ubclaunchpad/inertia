package client

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

var (
	// NoInertiaRemote is used to warn about missing inertia remote
	NoInertiaRemote = "No inertia remote"
)

// Config represents the current projects configuration.
type Config struct {
	Version   string       `toml:"inertia"`
	Project   string       `toml:"project-name"`
	BuildType string       `toml:"build-type"`
	remotes   []*RemoteVPS `toml:"remote"`
}

// NewConfig sets up Inertia configuration with given properties
func NewConfig(version, project, buildType string) *Config {
	return &Config{
		Version:   version,
		Project:   project,
		BuildType: buildType,
		remotes:   make([]*RemoteVPS, 0),
	}
}

// Write writes configuration to Inertia config file at path. Optionally
// takes io.Writers.
func (config *Config) Write(path string, writers ...io.Writer) error {
	var writer io.Writer

	// If io.Writers are given, attach all writers
	if len(writers) != 0 {
		writer = io.MultiWriter(writers...)
	}

	// If path is given, attach file writer
	if path != "" {
		w, err := os.OpenFile(path, os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		// Overwrite file if file exists
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			ioutil.WriteFile(path, []byte(""), 0644)
		}
		writer = io.MultiWriter(writer, w)
	}

	// Write configuration to writers
	encoder := toml.NewEncoder(writer)
	return encoder.Encode(config)
}

// GetRemotes returns all remotes attached to this configuration
func (config *Config) GetRemotes() []*RemoteVPS {
	return config.remotes
}

// GetRemote retrieves a remote by name
func (config *Config) GetRemote(name string) (*RemoteVPS, bool) {
	for _, remote := range config.remotes {
		if remote.Name == name {
			return remote, true
		}
	}
	return nil, false
}

// AddRemote adds a remote to configuration
func (config *Config) AddRemote(remote *RemoteVPS) {
	config.remotes = append(config.remotes, remote)
}

// RemoveRemote removes remote with given name
func (config *Config) RemoveRemote(name string) bool {
	for index, remote := range config.remotes {
		if remote.Name == name {
			remote = nil
			config.remotes = append(config.remotes[:index], config.remotes[index+1:]...)
			return true
		}
	}
	return false
}
