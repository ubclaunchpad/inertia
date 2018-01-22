// Copyright © 2017 UBC Launch Pad team@ubclaunchpad.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ubclaunchpad/inertia/common"
)

// Config represents the current projects configuration.
type Config struct {
	CurrentRemoteName string     `json:"name"`
	CurrentRemoteVPS  *RemoteVPS `json:"remote"`
	DaemonAPIToken    string     `json:"token"`
	io.Writer         `json:"-"`
}

const (
	// NoInertiaRemote is used to warn about missing inertia remote
	NoInertiaRemote  = "No inertia remote"
	configFolderName = ".inertia"
	configFileName   = "config.json"
)

// InitializeInertiaProject creates the inertia config folder and
// returns an error if we're not in a git project.
func InitializeInertiaProject() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = common.CheckForGit(cwd)
	if err != nil {
		return err
	}
	err = common.CheckForDockerCompose(cwd)
	if err != nil {
		return err
	}

	return createConfigDirectory()
}

// createConfigDirectory returns an error if the config directory
// already exists (the project is already initialized).
func createConfigDirectory() error {
	configDirPath, err := getProjectConfigFolderPath()
	if err != nil {
		return err
	}

	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	_, dirErr := os.Stat(configDirPath)
	s, fileErr := os.Stat(configFilePath)

	// Check if everything already exists.
	if s != nil {
		return errors.New("inertia already properly configured in this folder")
	}

	// Something doesn't exist.
	if os.IsNotExist(dirErr) {
		os.Mkdir(configDirPath, os.ModePerm)
	}

	// Directory exists. Make sure JSON exists.
	if os.IsNotExist(fileErr) {
		config := Config{
			CurrentRemoteName: NoInertiaRemote,
			CurrentRemoteVPS:  &RemoteVPS{},
			DaemonAPIToken:    "",
		}

		path, err := getConfigFilePath()
		if err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		writer, err := os.OpenFile(configFilePath, os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		config.Writer = writer
		defer f.Close()
		config.Write()
	}

	return nil
}

// Write writes configuration to JSON file in .inertia folder.
func (config *Config) Write() (int, error) {
	inertiaJSON, err := json.Marshal(config)
	if err != nil {
		return -1, err
	}

	return config.Writer.Write(inertiaJSON)
}

// GetProjectConfigFromDisk returns the current project's configuration.
// If an .inertia folder is not found, it returns an error.
func GetProjectConfigFromDisk() (*Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	raw, err := ioutil.ReadFile(configFilePath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("config file doesnt exist, try inertia init")
		}
		return nil, err
	}

	var result Config
	err = json.Unmarshal(raw, &result)
	if err != nil {
		return nil, err
	}

	// Add writer to object for writing/testing.
	result.Writer, err = os.OpenFile(configFilePath, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &result, err
}

// getProjectConfigFolderPath gets the absolute location of the project
// configuration folder.
func getProjectConfigFolderPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(cwd, configFolderName), nil
}

// getConfigFilePath returns the absolute path of the config JSON
// file.
func getConfigFilePath() (string, error) {
	path, err := getProjectConfigFolderPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, configFileName), nil
}
