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

func GetInertiaConfig() (*cfg.Inertia, error) {
	raw, err := ioutil.ReadFile(InertiaDir())
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

func SaveRemote(name string, remote *cfg.Remote) error {
	inertia, err := GetInertiaConfig()
	if err != nil {
		return err
	}
	if !inertia.AddRemote(name, *remote) {
		return fmt.Errorf("could not update remote '%s'", name)
	}
	return Write(InertiaDir(), inertia)
}

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
