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

const inertiaGlobalName = "inertia.global"

// InertiaDir gets the path to the directory where global Inertia configuration
// is stored
func InertiaDir() string {
	if os.Getenv("INERTIA_PATH") != "" {
		return os.Getenv("INERTIA_PATH")
	}
	home, err := GetHomePath()
	if err != nil {
		return "/.inertia"
	}
	return filepath.Join(home, ".inertia")
}

// InertiaConfigPath gets the path to global Inertia configuration
func InertiaConfigPath() string { return filepath.Join(InertiaDir(), inertiaGlobalName) }

// GetInertiaConfig retrieves global Inertia configuration
func GetInertiaConfig() (*cfg.Inertia, error) {
	raw, err := ioutil.ReadFile(InertiaConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("config file doesn't exist - try running inertia config init")
		}
		return nil, err
	}

	var inertia cfg.Inertia
	if err = toml.Unmarshal(raw, &inertia); err != nil {
		return nil, err
	}
	return &inertia, nil
}

// SaveRemote adds or updates the given remote in the global Inertia configuration
// file.
func SaveRemote(remote *cfg.Remote) error {
	inertia, err := GetInertiaConfig()
	if err != nil {
		return err
	}
	inertia.SetRemote(*remote)
	return Write(InertiaConfigPath(), inertia)
}

// RemoveRemote deletes the named remote from the global Inertia configuration file.
func RemoveRemote(name string) error {
	inertia, err := GetInertiaConfig()
	if err != nil {
		return err
	}
	if !inertia.RemoveRemote(name) {
		return fmt.Errorf("failed to remove remote '%s'", name)
	}
	return Write(InertiaConfigPath(), inertia)
}

// GetProject retrieves the Inertia project configuration at the given path
func GetProject(path string) (*cfg.Project, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("config file doesn't exist - try running inertia init")
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
