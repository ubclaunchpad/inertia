package main

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

// createConfigFile returns an error if the config directory
// already exists (the project is already initialized).
func createConfigFile(version, buildType string) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	// Check if Inertia is already set up.
	s, fileErr := os.Stat(configFilePath)
	if s != nil {
		return errors.New("inertia already properly configured in this folder")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Directory exists. Make sure configuration file exists.
	if os.IsNotExist(fileErr) {
		config := client.Config{
			Project:   filepath.Base(cwd),
			Version:   version,
			BuildType: buildType,
			Remotes:   make([]*client.RemoteVPS, 0),
		}

		path, err := getConfigFilePath()
		if err != nil {
			return err
		}
		f, err := os.Create(path)
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
func initializeInertiaProject(version, buildType string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = common.CheckForGit(cwd)
	if err != nil {
		return err
	}

	return createConfigFile(version, buildType)
}

// getProjectConfigFromDisk returns the current project's configuration.
// If an .inertia folder is not found, it returns an error.
func getProjectConfigFromDisk() (*client.Config, string, error) {
	configFilePath, err := getConfigFilePath()
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

// getConfigFilePath returns the absolute path of the config file.
func getConfigFilePath() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, configFileName), nil
}

// getClient returns a local deployment setup
func getClient(name string) (*client.Client, error) {
	config, _, err := getProjectConfigFromDisk()
	if err != nil {
		return nil, err
	}

	client, found := config.NewClient(name)
	if !found {
		return nil, errors.New("Remote not found")
	}

	return client, nil
}
