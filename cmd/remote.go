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
	Long: `Remote is a low level command for interacting with this VPS
over SSH. Provides functionality such as adding new remotes, removing remotes,
bootstrapping the server for deployment, running install scripts such as
installing docker, starting the Inertia daemon and other low level configuration
of the VPS. Must run 'inertia init' in your repository before using.

Example:

inerta remote add gcloud 35.123.55.12 -i /Users/path/to/pem
inerta remote bootstrap gcloud
inerta remote status gcloud`,
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
				fmt.Printf("%s\n", config.CurrentRemoteName)
				fmt.Printf("%+v\n", *config.CurrentRemoteVPS)
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

// bootstrapCmd represents the remote add command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap [REMOTE]",
	Short: "Bootstrap the VPS for continuous deployment",
	Long: `Bootstrap the VPS for continuous deployment.
A URL will be provided to direct GitHub webhooks to, the daemon will
request access to the repository via a public key, and will listen
for updates to this repository's remote master branch.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, err := GetProjectConfigFromDisk()
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		port, _ := cmd.Flags().GetString("port")
		if args[0] != config.CurrentRemoteName {
			println("No such remote " + args[0])
			println("Inertia currently supports one remote per repository")
			println("Run `inertia remote -v' to see what remote is available")
			os.Exit(1)
		}
		config.CurrentRemoteVPS.Bootstrap(args[0], port)
	},
}

// statusCmd represents the remote add command
var statusCmd = &cobra.Command{
	Use:   "status [REMOTE]",
	Short: "Query the status of a remote instance",
	Long: `Query the remote VPS for connectivity, daemon
behaviour, and other information.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, err := GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		if args[0] != config.CurrentRemoteName {
			println("No such remote " + args[0])
			println("Inertia currently supports one remote per repository")
			println("Run `inertia remote -v' to see what remote is available")
			os.Exit(1)
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

		fmt.Printf("Remote instance '%s' accepting requests at %s\n",
			config.CurrentRemoteName, host)
	},
}

func init() {
	RootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(addCmd)
	remoteCmd.AddCommand(bootstrapCmd)
	remoteCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	remoteCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	addCmd.Flags().StringP("user", "u", "root", "User for SSH access")
	addCmd.Flags().StringP("identity", "i", "$HOME/.ssh/id_rsa", "PEM file location")
	bootstrapCmd.Flags().StringP("port", "p", defaultDaemonPort,
		"The port for the daemon to listen on")
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

// Bootstrap configures a remote vps for continuous deployment
// by installing docker, starting the daemon and building a
// public-private key-pair. It outputs configuration information
// for the user.
func (remote *RemoteVPS) Bootstrap(name, daemonPort string) error {
	println("Bootstrapping remote " + name)

	println("Installing docker")
	err := remote.InstallDocker()
	if err != nil {
		return err
	}

	println("Starting daemon")
	err = remote.DaemonUp(daemonPort)
	if err != nil {
		return err
	}

	println("Building deploy key")
	pub, err := remote.KeyGen()
	if err != nil {
		return err
	}

	println()
	println("Daemon running on instance")

	// Output deploy key to user.
	println("GitHub Deploy Key (add here https://www.github.com/<your_repo>/settings/hooks/new): ")
	println(pub.String())

	// Output Webhook url to user.
	println("GitHub WebHook URL (add here https://www.github.com/<your_repo>/settings/keys/new): ")
	println("http://" + remote.IP + ":" + daemonPort)
	println("Github WebHook Secret: " + defaultSecret)

	println()

	println("Inertia daemon successfully deployed, add webhook url and deploy key to enable it.")
	fmt.Printf("Then run `inertia deploy %s' to deploy your application.\n", name)

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
