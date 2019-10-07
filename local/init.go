package local

import (
	"errors"
	"os"

	"github.com/ubclaunchpad/inertia/cfg"
)

// Initialize sets up Inertia configuration
func Initialize() (*cfg.Remotes, error) {
	var inertiaPath = InertiaDir()
	os.MkdirAll(inertiaPath, os.ModePerm)
	var remotes = cfg.NewRemotesConfig()
	return remotes, Write(InertiaRemotesPath(), remotes)
}

// InitProject creates the inertia config file and returns an error if Inertia
// is already configured
func InitProject(path, name, host string, defaultProfile cfg.Profile) error {
	if s, _ := os.Stat(path); s != nil {
		return errors.New("inertia is already properly configured in this directory")
	}

	var project = cfg.NewProject(name, host)
	defaultProfile.Name = "default"
	project.SetProfile(defaultProfile)

	return Write(path, project)
}
