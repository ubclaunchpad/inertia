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

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// RemoteVPS holds access to a remote instance.
type RemoteVPS struct {
	User string
	IP   string
	PEM  string
}

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

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "inertia",
	Short: "Inertia is a continuous-deployment scaffold",
	Long: `Inertia provides a continuous-deployment scaffold for applications.
Initialization involves preparing a server to run an application, then
activating a daemon which will continously update the production server
with new releases as they become available in the project's repository.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// CheckForGit raises an error if we're not in a git repository.
func CheckForGit() {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

	// Capture result.
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Output should be "true\n".
	if err != nil || string(out.Bytes()) != "true\n" {
		log.Fatal("This does not appear to be a git repository.")
	}
}

// GetProjectConfigFromDisk returns the current projects configuration.
// If an .inertia folder is not found, it creates one and
// adds the configuration JSON with default settings.
func GetProjectConfigFromDisk() *Config {
	CheckForGit()

	// If ConfigDir is missing, make it.
	configDirPath := GetProjectConfigFolderPath()
	if _, err := os.Stat(configDirPath); os.IsNotExist(err) {
		os.Mkdir(configDirPath, os.ModePerm)
		config := Config{
			CurrentRemoteName: noInertiaRemote,
			CurrentRemoteVPS:  RemoteVPS{},
		}
		config.Write()
	}

	configFilePath := filepath.Join(GetProjectConfigFolderPath(), configFileName)
	raw, err := ioutil.ReadFile(configFilePath)

	if err != nil {
		log.Fatal(err)
	}

	var result Config
	json.Unmarshal(raw, &result)

	return &result
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

// Write writes configuration to JSON file in .inertia folder.
func (config *Config) Write() {
	inertiaJSON, err := json.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Join(GetProjectConfigFolderPath(), configFileName)
	err = ioutil.WriteFile(path, inertiaJSON, 0644)
}
