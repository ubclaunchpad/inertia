package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/common"
)

const configFileName = ".inertia.toml"

// setProperty takes a struct pointer and searches for its "toml" tag with a search key
// and set property value with the tag
func setProperty(name string, value string, structObject interface{}) bool {
	val := reflect.ValueOf(structObject)

	if val.Kind() != reflect.Ptr {
		return false
	}
	structVal := val.Elem()
	for i := 0; i < structVal.NumField(); i++ {
		valueField := structVal.Field(i)
		typeField := structVal.Type().Field(i)
		if typeField.Tag.Get("toml") == name {
			if valueField.IsValid() && valueField.CanSet() && valueField.Kind() == reflect.String {
				valueField.SetString(value)
				return true
			}
		}
	}
	return false
}

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

	// If file does not exist, create new configuration file.
	if os.IsNotExist(fileErr) {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		config := client.NewConfig(version, filepath.Base(cwd), buildType)

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

	client, found := client.NewClient(name, config)
	if !found {
		return nil, errors.New("Remote not found")
	}

	return client, nil
}
