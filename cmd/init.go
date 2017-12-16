// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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

package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Config represents the current projects configuration.
type Config struct {
	CurrentRemoteName string    `json:"name"`
	CurrentRemoteVPS  RemoteVPS `json:"remote"`
}

var (
	configFolderName = ".inertia"
	configFileName   = "config.json"
	noInertiaRemote  = "No inertia remote"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := InitializeInertiaProject()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// InitializeInertiaProject creates the inertia config folder and
// returns an error if we're not in a git project.
func InitializeInertiaProject() error {
	err := CheckForGit()
	if err != nil {
		return err
	}

	err = CreateConfigDirectory()
	if err != nil {
		return err
	}

	return nil
}

// CreateConfigDirectory returns an error if the config directory
// already exists (the project is already initialized).
func CreateConfigDirectory() error {
	// Check if directory exists.
	configDirPath := GetProjectConfigFolderPath()
	configFilePath := GetConfigFilePath()

	_, dirErr := os.Stat(configDirPath)
	_, fileErr := os.Stat(configFilePath)

	// Check if everything already exists.
	if os.IsExist(dirErr) && os.IsExist(fileErr) {
		return errors.New("inertia already properly configured in this folder")
	}

	// Something doesn't exist.
	if os.IsNotExist(dirErr) {
		os.Mkdir(configDirPath, os.ModePerm)
	}

	// Directory exists. Make sure JSON exists.
	if os.IsNotExist(fileErr) {
		config := Config{
			CurrentRemoteName: noInertiaRemote,
			CurrentRemoteVPS:  RemoteVPS{},
		}
		config.Write()
	}

	return nil
}

// CheckForGit returns an error if we're not in a git repository.
func CheckForGit() error {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

	// Capture result.
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return err
	}

	// Output should be "true\n".
	if string(out.Bytes()) != "true\n" {
		return errors.New("this does not appear to be a git repository")
	}

	return nil
}

// GetProjectConfigFromDisk returns the current projects configuration.
// If an .inertia folder is not found, it returns an error.
func GetProjectConfigFromDisk() (*Config, error) {
	configFilePath := GetConfigFilePath()
	raw, err := ioutil.ReadFile(configFilePath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("config file doesnt exist, try inertia init")
		}
		log.Fatal(err)
	}

	var result Config
	json.Unmarshal(raw, &result)

	return &result, err
}

// GetProjectConfigFolderPath gets the absolute location of the project
// configuration folder.
func GetProjectConfigFolderPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	return filepath.Join(cwd, configFolderName)
}

// GetConfigFilePath returns the absolute path of the config JSON
// file.
func GetConfigFilePath() string {
	return filepath.Join(GetProjectConfigFolderPath(), configFileName)
}

// Write writes configuration to JSON file in .inertia folder.
func (config *Config) Write() {
	inertiaJSON, err := json.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}

	path := GetConfigFilePath()
	err = ioutil.WriteFile(path, inertiaJSON, 0644)
}
