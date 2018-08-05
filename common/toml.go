package common

import (
	"errors"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// InertiaProject represents inertia.toml, which contains Inertia's persistent
// project configurations. This file should be committed.
type InertiaProject struct {
	Version    *string      `toml:"version"`
	Project    *string      `toml:"project-name"`
	Build      InertiaBuild `toml:"build"`
	Repository InertiaRepo  `toml:"repository"`
}

// InertiaRepo represents general repository settings
type InertiaRepo struct {
	RemoteURL *string `toml:"remote-url"`
}

// InertiaBuild represents build settings
type InertiaBuild struct {
	Type       *string `toml:"type"`
	ConfigPath *string `toml:"config"`
}

// ReadProjectConfig reads project configuration from given filepath
func ReadProjectConfig(filepath string) (InertiaProject, error) {
	proj := InertiaProject{}
	projectBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return proj, errors.New("[WARNING] no inertia configuration found")
	}
	err = toml.Unmarshal(projectBytes, &proj)
	if err != nil {
		return proj, errors.New("[WARNING] inertia configuration invalid: " + err.Error())
	}
	return proj, nil
}

// StrDeref gets string value or empty string
func StrDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
