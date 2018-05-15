package client

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
)

var (
	// NoInertiaRemote is used to warn about missing inertia remote
	NoInertiaRemote = "No inertia remote"
)

// Config represents the current projects configuration.
type Config struct {
	Version   string       `toml:"version"`
	Project   string       `toml:"project-name"`
	BuildType string       `toml:"build-type"`
	Remotes   []*RemoteVPS `toml:"remote"`
}

// NewConfig sets up Inertia configuration with given properties
func NewConfig(version, project, buildType string) *Config {
	return &Config{
		Version:   version,
		Project:   project,
		BuildType: buildType,
		Remotes:   make([]*RemoteVPS, 0),
	}
}

// Write writes configuration to Inertia config file at path. Optionally
// takes io.Writers.
func (config *Config) Write(path string, writers ...io.Writer) error {
	if len(writers) == 0 && path == "" {
		return errors.New("nothing to write to")
	}

	var writer io.Writer

	// If io.Writers are given, attach all writers
	if len(writers) > 0 {
		writer = io.MultiWriter(writers...)
	}

	// If path is given, attach file writer
	if path != "" {
		w, err := os.OpenFile(path, os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}

		// Overwrite file if file exists
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			ioutil.WriteFile(path, []byte(""), 0644)
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

// SetProperty takes a struct pointer and searches for its "toml" tag with a search key
// and set property value with the tag
func SetProperty(name string, value string, structObject interface{}) bool {
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
