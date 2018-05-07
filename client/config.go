package client

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

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
	Version   string       `toml:"inertia"`
	Project   string       `toml:"project-name"`
	BuildType string       `toml:"build-type"`
	Remotes   []*RemoteVPS `toml:"remote"`
	Writer    io.Writer    `toml:"-"`
}

// Write writes configuration to Inertia config file.
func (config *Config) Write() error {
	if config.Writer == nil {
		return nil
	}
	path, err := GetConfigFilePath()
	if err != nil {
		return err
	}
	// Overwrite file if file exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		ioutil.WriteFile(path, []byte(""), 0644)
	}
	// Write configuration to file
	encoder := toml.NewEncoder(config.Writer)
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
func SetProperty(name string, value string,structObject interface{}) bool {
	val := reflect.ValueOf(structObject)

	if val.Kind() != reflect.Ptr {
		println("Interal error: invalid interface.Must be a ptr to struct")
		return false
	}
	structVal := val.Elem()
	for i := 0; i < structVal.NumField(); i++ {
		valueField := structVal.Field(i)
		typeField := structVal.Type().Field(i)
		if typeField.Tag.Get("toml") == name{
			if valueField.IsValid() {
				if valueField.CanSet() {
					if valueField.Kind() == reflect.String {
						valueField.SetString(value)
						return true
					}
				}
			}
		}
	}
	return false
}


// InitializeInertiaProject creates the inertia config folder and
// returns an error if we're not in a git project.
func InitializeInertiaProject(version, buildType string) error {
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

// createConfigFile returns an error if the config directory
// already exists (the project is already initialized).
func createConfigFile(version, buildType string) error {
	configFilePath, err := GetConfigFilePath()
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
		config := Config{
			Project:   filepath.Base(cwd),
			Version:   version,
			BuildType: buildType,
			Remotes:   make([]*RemoteVPS, 0),
		}

		path, err := GetConfigFilePath()
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
	configFilePath, err := GetConfigFilePath()
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

// GetConfigFilePath returns the absolute path of the config file.
func GetConfigFilePath() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, configFileName), nil
}

