package cfg

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/ubclaunchpad/inertia/common"
)

var (
	// NoInertiaRemote is used to warn about missing inertia remote
	NoInertiaRemote = "no inertia remote"

	// NoInertiaConfig is used to warn about missing inertia configuration
	NoInertiaConfig = "no inertia configuration found - try running 'inertia init'"
)

// Config represents the current project's configuration.
type Config struct {
	Version       string
	Project       string
	BuildType     string
	BuildFilePath string
	RemoteURL     string

	remotes map[string]*RemoteVPS
}

// NewConfig sets up Inertia configuration with given properties
func NewConfig(version, project, buildType, buildFilePath, remoteURL string) *Config {
	cfg := &Config{
		Version:   version,
		Project:   project,
		BuildType: buildType,
		RemoteURL: remoteURL,
		remotes:   make(map[string]*RemoteVPS),
	}
	if buildFilePath != "" {
		cfg.BuildFilePath = buildFilePath
	}
	return cfg
}

// NewConfigFromFiles loads configuration from given filepaths
func NewConfigFromFiles(projectConfigPath string, remoteConfigPath string) (*Config, error) {
	// Attempt to read files
	projectBytes, err := ioutil.ReadFile(projectConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	remotesBytes, err := ioutil.ReadFile(remoteConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	var (
		project = &common.InertiaProject{}
		remotes = &InertiaRemotes{}
	)

	// If both files are present, construct config using both
	if projectBytes != nil && remotesBytes != nil {
		err = toml.Unmarshal(projectBytes, project)
		if err != nil {
			return nil, fmt.Errorf("project config error: %s", err.Error())
		}
		err = toml.Unmarshal(remotesBytes, remotes)
		if err != nil {
			return nil, fmt.Errorf("remotes config error: %s", err.Error())
		}
		return NewConfigFromTOML(*project, *remotes)
	}

	// If only project is present, construct config using project
	if projectBytes != nil && remotesBytes == nil {
		err = toml.Unmarshal(projectBytes, project)
		if err != nil {
			return nil, fmt.Errorf("project config error: %s", err.Error())
		}
		return NewConfigFromTOML(*project, InertiaRemotes{Version: project.Version})
	}

	// If only remotes is present, construct config using remotes
	if projectBytes == nil && remotesBytes != nil {
		err = toml.Unmarshal(remotesBytes, remotes)
		if err != nil {
			return nil, fmt.Errorf("remotes config error: %s", err.Error())
		}
		return NewConfigFromTOML(common.InertiaProject{
			Version: remotes.Version,
		}, *remotes)
	}

	return nil, errors.New(NoInertiaConfig)
}

// NewConfigFromTOML loads configuration from TOML format structs
func NewConfigFromTOML(project common.InertiaProject, remotes InertiaRemotes) (*Config, error) {
	// Set remote defaults
	if remotes.Remotes == nil {
		r := make(map[string]*RemoteVPS)
		remotes.Remotes = &r
	}
	if remotes.Version == nil {
		remotes.Version = project.Version
	}

	// Check all is g
	if *project.Version != *remotes.Version {
		return nil, fmt.Errorf("mismatching versions %s and %s", *project.Version, *remotes.Version)
	}

	// Generate configuration
	return &Config{
		Version:       *project.Version,
		Project:       *project.Project,
		BuildType:     *project.BuildType,
		BuildFilePath: *project.BuildFilePath,
		RemoteURL:     *project.Repository.RemoteURL,
		remotes:       *remotes.Remotes,
	}, nil
}

// GetRemotes retrieves a list of all remotes
func (config *Config) GetRemotes() []*RemoteVPS {
	remotes := make([]*RemoteVPS, 0)
	for name, remote := range config.remotes {
		// Set name
		remote.Name = name
		remotes = append(remotes, remote)
	}
	return remotes
}

// GetRemote retrieves a remote by name
func (config *Config) GetRemote(name string) (*RemoteVPS, bool) {
	for n, remote := range config.remotes {
		if n == name {
			// Set name
			remote.Name = n
			return remote, true
		}
	}
	return nil, false
}

// AddRemote adds a remote to configuration
func (config *Config) AddRemote(remote *RemoteVPS) bool {
	_, ok := config.remotes[remote.Name]
	if ok {
		return false
	}
	config.remotes[remote.Name] = remote
	return true
}

// RemoveRemote removes remote with given name
func (config *Config) RemoveRemote(name string) bool {
	_, ok := config.remotes[name]
	if !ok {
		return false
	}
	delete(config.remotes, name)
	return true
}

// GetProjectConfig gets project configuration
func (config *Config) GetProjectConfig() *common.InertiaProject {
	return &common.InertiaProject{
		Version:       &config.Version,
		Project:       &config.Project,
		BuildType:     &config.BuildType,
		BuildFilePath: &config.BuildFilePath,
		Repository:    &common.InertiaRepo{&config.RemoteURL}}
}

// WriteProjectConfig writes Inertia project configuration. This file should be
// committed.
func (config *Config) WriteProjectConfig(filePath string, writers ...io.Writer) error {
	toml := config.GetProjectConfig()
	return config.write(filePath, toml, writers...)
}

// WriteRemoteConfig writes Inertia remote configuration. This file should NOT
// be committed.
func (config *Config) WriteRemoteConfig(filePath string, writers ...io.Writer) error {
	toml := InertiaRemotes{&config.Version, &config.remotes}
	return config.write(filePath, toml, writers...)
}

// write writes configuration to Inertia config file at path. Optionally
// takes io.Writers.
func (config *Config) write(filePath string, contents interface{}, writers ...io.Writer) error {
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
	return encoder.Encode(contents)
}
