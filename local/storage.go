package local

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ubclaunchpad/inertia/cfg"
)

const inertiaRemotesFile = "inertia.remotes"

// InertiaDir gets the path to the directory where global Inertia configuration
// is stored
func InertiaDir() string {
	if os.Getenv("INERTIA_PATH") != "" {
		return os.Getenv("INERTIA_PATH")
	}
	confDir, err := os.UserConfigDir()
	if err != nil {
		return "/inertia"
	}
	return filepath.Join(confDir, "inertia")
}

// InertiaRemotesPath gets the path to global Inertia configuration
func InertiaRemotesPath() string { return filepath.Join(InertiaDir(), inertiaRemotesFile) }

// GetRemotes retrieves global Inertia remotes configuration
func GetRemotes() (*cfg.Remotes, error) {
	raw, err := ioutil.ReadFile(InertiaRemotesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("global config file doesn't exist - try running 'inertia init --global'")
		}
		return nil, err
	}

	var remotes cfg.Remotes
	if err = toml.Unmarshal(raw, &remotes); err != nil {
		return nil, err
	}
	return &remotes, nil
}

// SaveRemote adds or updates the given remote in the global Inertia configuration
// file.
func SaveRemote(remote *cfg.Remote) error {
	remotes, err := GetRemotes()
	if err != nil {
		return err
	}
	remotes.SetRemote(*remote)
	return Write(InertiaRemotesPath(), remotes)
}

// RemoveRemote deletes the named remote from the global Inertia configuration file.
func RemoveRemote(name string) error {
	remotes, err := GetRemotes()
	if err != nil {
		return err
	}
	if !remotes.RemoveRemote(name) {
		return fmt.Errorf("failed to remove remote '%s'", name)
	}
	return Write(InertiaRemotesPath(), remotes)
}

// GetProject retrieves the Inertia project configuration at the given path
func GetProject(path string) (*cfg.Project, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("project config file doesn't exist - try running 'inertia init'")
		}
		return nil, err
	}

	var project cfg.Project
	if err = toml.Unmarshal(raw, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// Write saves the given data to the given path and/or writers
func Write(path string, data interface{}, writers ...io.Writer) error {
	if len(writers) == 0 && path == "" {
		return errors.New("nothing to write to")
	}

	// If io.Writers are given, attach all writers
	var writer io.Writer
	if len(writers) > 0 {
		writer = io.MultiWriter(writers...)
	}

	// If path is given, attach file writer
	if path != "" {
		w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}

		// Overwrite file if file exists
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			ioutil.WriteFile(path, []byte(""), os.ModePerm)
		} else if err != nil {
			return err
		}

		// Set writer
		if writer != nil {
			writer = io.MultiWriter(writer, w)
		} else {
			writer = w
		}
	} else {
		writer = os.Stdout
	}

	// Write configuration to writers
	encoder := toml.NewEncoder(writer)
	return encoder.Encode(data)
}

// SaveKey writes a key to given path
func SaveKey(keyMaterial string, path string) error {
	return ioutil.WriteFile(path, []byte(keyMaterial), 0400)
}
