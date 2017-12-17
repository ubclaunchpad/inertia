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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var execCommand = exec.Command

// RemoteVPS holds access to a remote instance.
type RemoteVPS struct {
	User string
	IP   string
	PEM  string
	Port string
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
		// Ensure project initialized.
		_, err := GetProjectConfigFromDisk()
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		user, _ := cmd.Flags().GetString("user")
		pemLoc, _ := cmd.Flags().GetString("identity")
		AddNewRemote(args[0], args[1], user, pemLoc)
	},
}

// statusCmd represents the remote add command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Query the status of a remote instance",
	Long: `Query the remote VPS for connectivity, daemon
behaviour, and other information.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		host := "http://" + config.CurrentRemoteVPS.GetIPAndPort()
		resp, err := http.Get(host)
		if err != nil {
			println("Could not connect to daemon")
			println("Try running inertia deploy")
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if string(body) != okResp {
			println("Could not connect to daemon")
			println("Try running inertia deploy")
			return
		}

		println("Remote instance accepting requests at " + host)
	},
}

func init() {
	RootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(addCmd)
	remoteCmd.AddCommand(statusCmd)

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
	config.CurrentRemoteVPS = &RemoteVPS{
		IP:   IP,
		User: user,
		PEM:  pemLoc,
	}

	f, err := GetConfigFile()
	defer f.Close()
	config.Write(f)

	println("Remote '" + name + "' added.")

	return nil
}

// GetHost creates the user@IP string.
func (remote *RemoteVPS) GetHost() string {
	return remote.User + "@" + remote.IP
}

// GetIPAndPort creates the IP:Port string.
func (remote *RemoteVPS) GetIPAndPort() string {
	return remote.IP + ":" + remote.Port
}

// RunSSHCommand runs a command remotely.
func (remote *RemoteVPS) RunSSHCommand(remoteCmd string) (
	*bytes.Buffer, *bytes.Buffer, error) {
	cmd := execCommand("ssh", "-i", remote.PEM, "-t", remote.GetHost(), remoteCmd)

	// Capture result.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return &stdout, &stderr, err
}

// InstallDocker installs docker on a remote vps.
func (remote *RemoteVPS) InstallDocker() error {
	// Collect assets (docker shell script)
	installDockerSh, err := Asset("cmd/bootstrap/docker.sh")
	if err != nil {
		return err
	}

	// Install docker.
	_, stderr, err := remote.RunSSHCommand(string(installDockerSh))
	if err != nil {
		log.Println(stderr)
		return err
	}

	return nil
}

// DaemonUp brings the daemon up on the remote instance.
func (remote *RemoteVPS) DaemonUp(daemonPort string) error {
	// Collect assets (deamon-up shell script)
	daemonCmd, err := Asset("cmd/bootstrap/daemon-up.sh")
	if err != nil {
		return err
	}

	// Run inertia daemon.
	daemonCmdStr := fmt.Sprintf(string(daemonCmd), daemonPort)
	_, stderr, err := remote.RunSSHCommand(daemonCmdStr)
	if err != nil {
		log.Println(stderr)
		return err
	}

	return nil
}

// KeyGen creates a public-private key-pair on the remote vps
// and returns the public key.
func (remote *RemoteVPS) KeyGen() (*bytes.Buffer, error) {
	// Collect assets (keygen shell script)
	keygenSh, err := Asset("cmd/bootstrap/keygen.sh")
	if err != nil {
		return nil, err
	}

	// Create deploy key.
	result, stderr, err := remote.RunSSHCommand(string(keygenSh))

	if err != nil {
		log.Println(stderr)
		return nil, err
	}

	return result, nil
}

// DaemonDown brings the daemon down on the remote instance
func (remote *RemoteVPS) DaemonDown() error {
	// Collect assets (deamon-up shell script)
	daemonCmd, err := Asset("cmd/bootstrap/daemon-down.sh")
	if err != nil {
		return err
	}

	_, stderr, err := remote.RunSSHCommand(string(daemonCmd))
	if err != nil {
		log.Println(stderr)
		return err
	}

	return nil
}
