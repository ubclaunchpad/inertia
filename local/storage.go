package local

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/BurntSushi/toml"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client"
)

const configFileName = "inertia.toml"

// InitializeInertiaProject creates the inertia config folder and
// returns an error if we're not in a git project.
func InitializeInertiaProject(version, buildType, buildFilePath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = checkForGit(cwd)
	if err != nil {
		return err
	}

	return createConfigFile(version, buildType, buildFilePath)
}

// createConfigFile returns an error if the config directory
// already exists (the project is already initialized).
func createConfigFile(version, buildType, buildFilePath string) error {
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
		config := cfg.NewConfig(version, filepath.Base(cwd), buildType, buildFilePath)

		f, err := os.Create(configFilePath)
		if err != nil {
			return err
		}
		defer f.Close()
		config.Write(configFilePath)
	}

	return nil
}

// GetProjectConfigFromDisk returns the current project's configuration.
// If an .inertia folder is not found, it returns an error.
func GetProjectConfigFromDisk() (*cfg.Config, string, error) {
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

	var cfg cfg.Config
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
func GetClient(name string, cmd ...*cobra.Command) (*client.Client, error) {
	config, _, err := GetProjectConfigFromDisk()
	if err != nil {
		return nil, err
	}

	client, found := client.NewClient(name, config)
	if !found {
		return nil, errors.New("Remote not found")
	}

	if len(cmd) == 1 && cmd[0] != nil {
		verify, err := cmd[0].Flags().GetBool("verify-ssl")
		if err != nil {
			return nil, err
		}
		client.SetSSLVerification(verify)
	}

	return client, nil
}
