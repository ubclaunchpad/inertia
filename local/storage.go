package local

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client"
)

// GetRemotesConfigFilePath retrieves path of file in global storage ($HOME/.inertia)
func GetRemotesConfigFilePath(projectName string) string {
	path := filepath.Join(".inertia", projectName+".remotes")
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("HOMEPATH"), path)
	}
	return filepath.Join(os.Getenv("HOME"), path)
}

// InitializeInertiaProject creates the inertia config folder and
// returns an error if we're not in a git project.
func InitializeInertiaProject(
	projectConfigPath, remoteConfigPath, version,
	buildType, buildFilePath, remoteURL string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = checkForGit(cwd)
	if err != nil {
		return err
	}

	return createConfigFile(projectConfigPath, remoteConfigPath, version, buildType, buildFilePath, remoteURL)
}

// createConfigFile returns an error if the config directory
// already exists (the project is already initialized).
func createConfigFile(projectConfigPath, remoteConfigPath, version, buildType, buildFilePath, remoteURL string) error {
	// Check if Inertia is already set up.
	s, fileErr := os.Stat(projectConfigPath)
	if s != nil {
		return errors.New("Inertia already properly configured in this folder")
	}

	// If file does not exist, create new configuration file.
	if os.IsNotExist(fileErr) {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		config := cfg.NewConfig(
			version, filepath.Base(cwd), buildType, buildFilePath, remoteURL)

		f, err := os.Create(projectConfigPath)
		if err != nil {
			return err
		}
		defer f.Close()
		config.WriteProjectConfig(projectConfigPath)
		config.WriteRemoteConfig(remoteConfigPath)
		return nil
	}

	return fileErr
}

// GetClient returns a local deployment setup. Returns a callback to write any
// changes made to the configuration.
func GetClient(name, projectConfigPath, remoteConfigPath string, cmd ...*cobra.Command) (*client.Client, func() error, error) {
	config, err := cfg.NewConfigFromFiles(projectConfigPath, remoteConfigPath)
	if err != nil {
		return nil, nil, err
	}

	client, found := client.NewClient(name, config, os.Stdout)
	if !found {
		return nil, nil, errors.New("Remote not found")
	}

	if len(cmd) == 1 && cmd[0] != nil {
		verify, err := cmd[0].Flags().GetBool("verify-ssl")
		if err != nil {
			return nil, nil, err
		}
		client.SetSSLVerification(verify)
	}

	return client, func() error {
		err = config.WriteProjectConfig(projectConfigPath)
		if err != nil {
			return err
		}
		return config.WriteRemoteConfig(remoteConfigPath)
	}, nil
}

// SaveKey writes a key to given path
func SaveKey(keyMaterial string, path string) error {
	return ioutil.WriteFile(path, []byte(keyMaterial), 0644)
}
