package local

import (
	"errors"
	"os"

	"github.com/ubclaunchpad/inertia/cfg"
)

// Init sets up Inertia configuration
func Init() (*cfg.Inertia, error) {
	var inertiaPath = InertiaDir()
	os.MkdirAll(inertiaPath, 0400)
	var inertia = cfg.NewInertiaConfig()
	return inertia, Write(InertiaConfigPath(), inertia)
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
