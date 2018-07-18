package common

import (
	"errors"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// InertiaProject represents inertia.toml, which contains Inertia's persistent
// project configurations. This file should be committed.
type InertiaProject struct {
	Version       *string      `toml:"version"`
	Project       *string      `toml:"project-name"`
	BuildType     *string      `toml:"build-type"`
	BuildFilePath *string      `toml:"build-file-path"`
	Repository    *InertiaRepo `toml:"repository"`
}

// InertiaRepo represents general repository settings
type InertiaRepo struct {
	RemoteURL *string `toml:"remote-url"`
}

// ReadProjectConfig reads project configuration from given filepath
func ReadProjectConfig(filepath string) (*InertiaProject, error) {
	projectBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.New("[WARNING] no inertia configuration found")
	}
	var proj *InertiaProject
	err = toml.Unmarshal(projectBytes, proj)
	if err != nil {
		return nil, errors.New("[WARNING] inertia configuration invalid: " + err.Error())
	}
	return proj, nil
}
