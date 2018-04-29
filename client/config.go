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
	Remotes   []*RemoteVPS `toml:"remote"`
}

// Write writes configuration to Inertia config file at path. Optionally
// takes io.Writers.
func (config *Config) Write(path string, writers ...io.Writer) error {
	var writer io.Writer
	if len(writers) == 0 {
		w, err := os.OpenFile(path, os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		writer = w
		// Overwrite file if file exists
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			ioutil.WriteFile(path, []byte(""), 0644)
		}
	} else {
		writer = io.MultiWriter(writers...)
	}

	// Write configuration to file
	encoder := toml.NewEncoder(writer)
	return encoder.Encode(config)
}

// GetRemote retrieves a remote by name
func (config *Config) GetRemote(name string) (*RemoteVPS, bool) {
	for _, remote := range config.Remotes {
		if remote.Name == name {
			return remote, true
		}
	}
	return nil, false
}

// AddRemote adds a remote to configuration
func (config *Config) AddRemote(remote *RemoteVPS) {
	config.Remotes = append(config.Remotes, remote)
}

// RemoveRemote removes remote with given name
func (config *Config) RemoveRemote(name string) bool {
	for index, remote := range config.Remotes {
		if remote.Name == name {
			remote = nil
			config.Remotes = append(config.Remotes[:index], config.Remotes[index+1:]...)
			return true
		}
	}
	return false
}
