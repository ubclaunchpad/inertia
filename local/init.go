package local

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/local/git"
)

// Init sets up Inertia configuration
func Init() (*cfg.Inertia, error) {
	var inertiaPath = InertiaDir()
	os.MkdirAll(inertiaPath, 0400)
	var configPath = filepath.Join(inertiaPath, "inertia.global")
	var inertia = &cfg.Inertia{
		Remotes: make(map[string]cfg.Remote),
	}
	return inertia, Write(configPath, inertia)
}

// InitProject creates the inertia config folder and returns an error if we're
// not in a git project.
func InitProject(path, name string, defaultProfile cfg.Profile) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err = git.IsRepo(cwd); err != nil {
		return fmt.Errorf("could not find git repository: %s", err.Error())
	}
	if s, _ := os.Stat(path); s != nil {
		return errors.New("inertia is already properly configured in this directory")
	}
	var project = cfg.NewProject(name)
	project.SetProfile("default", defaultProfile)

	return Write(path, project)
}
