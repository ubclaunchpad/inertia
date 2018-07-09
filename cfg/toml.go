package cfg

import (
	"errors"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// TOML formats are defined here, used by Config.Write()

// InertiaProject represents inertia.toml, which contains Inertia's persistent
// project configurations. This file should be committed.
type InertiaProject struct {
	Version       *string `toml:"version"`
	Project       *string `toml:"project-name"`
	BuildType     *string `toml:"build-type"`
	BuildFilePath *string `toml:"build-file-path"`
}

// InertiaRemotes represents inertia.remotes, which contains Inertia's runtime
// configuration for this project. This file should NOT be committed.
type InertiaRemotes struct {
	Version *string                `toml:"version"`
	Remotes *map[string]*RemoteVPS `toml:"remotes"`
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
