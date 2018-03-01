package client

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/ubclaunchpad/inertia/common"
)

var (
	// NoInertiaRemote is used to warn about missing inertia remote
	NoInertiaRemote = "No inertia remote"
	configFileName  = ".inertia.toml"
)

// Config represents the current projects configuration.
type Config struct {
	Project string       `toml:"project"`
	Remotes []*RemoteVPS `toml:"remote"`
	Writer  io.Writer    `toml:"-"`
}

// Write writes configuration to Inertia config file.
func (config *Config) Write() error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	// Overwrite file if file exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		ioutil.WriteFile(path, []byte(""), 0644)
	}
	// Write configuration to file
	encoder := toml.NewEncoder(config.Writer)
	encoder.Indent = "    "
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

// RemoveRemote removes remote at with given name
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

// InitializeInertiaProject creates the inertia config folder and
// returns an error if we're not in a git project.
func InitializeInertiaProject() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = common.CheckForGit(cwd)
	if err != nil {
		return err
	}
	err = common.CheckForDockerCompose(cwd)
	if err != nil {
		return err
	}

	return createConfigFile()
}

// createConfigFile returns an error if the config directory
// already exists (the project is already initialized).
func createConfigFile() error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	s, fileErr := os.Stat(configFilePath)

	// Check if everything already exists.
	if s != nil {
		return errors.New("inertia already properly configured in this folder")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Directory exists. Make sure JSON exists.
	if os.IsNotExist(fileErr) {
		config := Config{
			Project: filepath.Base(cwd),
			Remotes: make([]*RemoteVPS, 0),
		}

		path, err := getConfigFilePath()
		if err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		writer, err := os.OpenFile(configFilePath, os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		config.Writer = writer
		defer f.Close()
		config.Write()
	}

	return nil
}

// GetProjectConfigFromDisk returns the current project's configuration.
// If an .inertia folder is not found, it returns an error.
func GetProjectConfigFromDisk() (*Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	raw, err := ioutil.ReadFile(configFilePath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("config file doesnt exist, try inertia init")
		}
		return nil, err
	}

	var result Config
	err = toml.Unmarshal(raw, &result)
	if err != nil {
		return nil, err
	}

	// Add writer to object for writing/testing.
	result.Writer, err = os.OpenFile(configFilePath, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &result, err
}

// getConfigFilePath returns the absolute path of the config file.
func getConfigFilePath() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, configFileName), nil
}
