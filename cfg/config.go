package cfg

import (
	"errors"
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
	Version       string `toml:"version"`
	Project       string `toml:"project-name"`
	BuildType     string `toml:"build-type"`
	BuildFilePath string `toml:"build-file-path"`

	Remotes map[string]*RemoteVPS `toml:"remotes"`
}

// NewConfig sets up Inertia configuration with given properties
func NewConfig(version, project, buildType, buildFilePath string) *Config {
	cfg := &Config{
		Version:   version,
		Project:   project,
		BuildType: buildType,
		Remotes:   make(map[string]*RemoteVPS),
	}
	if buildFilePath != "" {
		cfg.BuildFilePath = buildFilePath
	}
	return cfg
}

// Write writes configuration to Inertia config file at path. Optionally
// takes io.Writers.
func (config *Config) Write(filePath string, writers ...io.Writer) error {
	if len(writers) == 0 && filePath == "" {
		return errors.New("nothing to write to")
	}

	var writer io.Writer

	// If io.Writers are given, attach all writers
	if len(writers) > 0 {
		writer = io.MultiWriter(writers...)
	}

	// If path is given, attach file writer
	if filePath != "" {
		w, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}

		// Overwrite file if file exists
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			ioutil.WriteFile(filePath, []byte(""), 0644)
		} else if err != nil {
			return err
		}

		// Set writer
		if writer != nil {
			writer = io.MultiWriter(writer, w)
		} else {
			writer = w
		}
	}

	// Write configuration to writers
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
func (config *Config) AddRemote(remote *RemoteVPS) bool {
	_, ok := config.Remotes[remote.Name]
	if ok {
		return false
	}
	config.Remotes[remote.Name] = remote
	return true
}

// RemoveRemote removes remote with given name
func (config *Config) RemoveRemote(name string) bool {
	_, ok := config.Remotes[name]
	if !ok {
		return false
	}
	delete(config.Remotes, name)
	return true
}
