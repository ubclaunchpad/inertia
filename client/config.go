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

// Config represents the current projects configuration.
type Config struct {
	Project   string                `toml:"project"`
	Remotes   map[string]*RemoteVPS `toml:"remotes"`
	io.Writer `toml:"-"`
}

var (
	// NoInertiaRemote is used to warn about missing inertia remote
	NoInertiaRemote = "No inertia remote"
	configFileName  = ".inertia.toml"
)

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

	return createConfigDirectory()
}

// createConfigDirectory returns an error if the config directory
// already exists (the project is already initialized).
func createConfigDirectory() error {
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
			Remotes: make(map[string]*RemoteVPS),
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

// Write writes configuration to Inertia config file.
func (config *Config) Write() error {
	encoder := toml.NewEncoder(config.Writer)
	return encoder.Encode(config)
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
