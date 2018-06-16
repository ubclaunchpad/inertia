package local

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/common"
)

const configFileName = ".inertia.toml"

// CreateConfigFile returns an error if the config directory
// already exists (the project is already initialized).
func CreateConfigFile(version, buildType, buildFilePath string) error {
	configFilePath, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	// Check if Inertia is already set up.
	s, fileErr := os.Stat(configFilePath)
	if s != nil {
		return errors.New("inertia already properly configured in this folder")
	}

	// If file does not exist, create new configuration file.
	if os.IsNotExist(fileErr) {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		config := client.NewConfig(version, filepath.Base(cwd), buildType, buildFilePath)

		f, err := os.Create(configFilePath)
		if err != nil {
			return err
		}
		defer f.Close()
		config.Write(configFilePath)
	}

	return nil
}

// InitializeInertiaProject creates the inertia config folder and
// returns an error if we're not in a git project.
func InitializeInertiaProject(version, buildType, buildFilePath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = common.CheckForGit(cwd)
	if err != nil {
		return err
	}

	return CreateConfigFile(version, buildType, buildFilePath)
}

// GetProjectConfigFromDisk returns the current project's configuration.
// If an .inertia folder is not found, it returns an error.
func GetProjectConfigFromDisk() (*client.Config, string, error) {
	configFilePath, err := GetConfigFilePath()
	if err != nil {
		return nil, "", err
	}

	raw, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, configFilePath, errors.New("config file doesnt exist, try inertia init")
		}
		return nil, configFilePath, err
	}

	var cfg client.Config
	err = toml.Unmarshal(raw, &cfg)
	if err != nil {
		return nil, configFilePath, err
	}

	return &cfg, configFilePath, err
}

// GetConfigFilePath returns the absolute path of the config file.
func GetConfigFilePath() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, configFileName), nil
}

// GetClient returns a local deployment setup
func GetClient(name string) (*client.Client, error) {
	config, _, err := GetProjectConfigFromDisk()
	if err != nil {
		return nil, err
	}

	client, found := client.NewClient(name, config)
	if !found {
		return nil, errors.New("Remote not found")
	}

	return client, nil
}
