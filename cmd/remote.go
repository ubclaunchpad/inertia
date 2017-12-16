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
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

// RemoteVPS holds access to a remote instance.
type RemoteVPS struct {
	User string
	IP   string
	PEM  string
}

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Configure the local settings for a remote VPS instance",
	Long: `Configure the local settings for a remote VPS instance. Requires
SSH access to the instance via a PEM file. Similar behaviour to
git remote`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		config, err := GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}
		if config.CurrentRemoteName == noInertiaRemote {
			println("No remote currently set.")
		} else {
			if verbose {
				fmt.Printf("%+v\n", config)
			} else {
				println(config.CurrentRemoteName)
			}
		}
	},
}

// addCmd represents the remote add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a reference to a remote VPS instance",
	Long: `Add a reference to a remote VPS instance. Requires 
information about the VPS including IP address, user and a PEM
file. Specify a VPS name and an IP address.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		user, _ := cmd.Flags().GetString("user")
		pemLoc, _ := cmd.Flags().GetString("identity")
		AddNewRemote(args[0], args[1], user, pemLoc)
	},
}

func init() {
	RootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	addCmd.Flags().StringP("user", "u", "root", "User for SSH access")
	addCmd.Flags().StringP("identity", "i", "$HOME/.ssh/id_rsa", "PEM file location")
	remoteCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}

// AddNewRemote adds a new remote to the project config file.
func AddNewRemote(name, IP, user, pemLoc string) error {
	// Just wipe configuration for MVP.
	config, err := GetProjectConfigFromDisk()
	if err != nil {
		return err
	}

	config.CurrentRemoteName = name
	config.CurrentRemoteVPS = RemoteVPS{
		IP:   IP,
		User: user,
		PEM:  pemLoc,
	}

	config.Write()
	println("Remote '" + name + "' added.")

	return nil
}

// GetHost creates the user@IP string.
func (remote *RemoteVPS) GetHost() string {
	return remote.User + "@" + remote.IP
}

// RunSSHCommand runs a command remotely.
func (remote *RemoteVPS) RunSSHCommand(remoteCmd string) (
	*bytes.Buffer, *bytes.Buffer, error) {
	cmd := exec.Command("ssh", "-i", remote.PEM, "-t", remote.GetHost(), remoteCmd)

	// Capture result.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return &stdout, &stderr, err
}
