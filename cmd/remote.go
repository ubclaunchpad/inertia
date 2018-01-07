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
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// RemoteVPS holds access to a remote instance.
type RemoteVPS struct {
	User string
	IP   string
	PEM  string
	Port string
}

// SSHSession can run remote commands over SSH
type SSHSession interface {
	Run(cmd string) (*bytes.Buffer, *bytes.Buffer, error)
}

// SSHRunner runs commands over SSH and captures results.
type SSHRunner struct {
	r *RemoteVPS
}

// Run runs a command remotely.
func (runner *SSHRunner) Run(cmd string) (*bytes.Buffer, *bytes.Buffer, error) {
	session, err := getSSHSession(runner.r.PEM, runner.r.IP, runner.r.User)
	if err != nil {
		return nil, nil, err
	}
	// Capture result.
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)
	return &stdout, &stderr, err
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
			log.WithError(err)
		}
		if config.CurrentRemoteName == noInertiaRemote {
			println("No remote currently set.")
		} else if verbose {
			fmt.Printf("%s\n", config.CurrentRemoteName)
			fmt.Printf("%+v\n", config.CurrentRemoteVPS)
		} else {
			println(config.CurrentRemoteName)
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
			log.WithError(err)
		}
		user, _ := cmd.Flags().GetString("user")
		pemLoc, _ := cmd.Flags().GetString("identity")
		port, _ := cmd.Flags().GetString("port")
		addNewRemote(args[0], args[1], user, pemLoc, port)
	},
}

// deployInitCmd represents the inertia [REMOTE] init command
var deployInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the VPS for continuous deployment",
	Long: `Initialize the VPS for continuous deployment.
This sets up everything you might need and brings the Inertia daemon
online on your remote.
A URL will be provided to direct GitHub webhooks to, the daemon will
request access to the repository via a public key, and will listen
for updates to this repository's remote master branch.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: chagne support correct remote based on which
		// cmd is calling this init, see "deploy.go"

		// Ensure project initialized.
		config, err := GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		session := &SSHRunner{r: config.CurrentRemoteVPS}
		err = config.CurrentRemoteVPS.Bootstrap(session, args[0], config)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// statusCmd represents the remote status command
var statusCmd = &cobra.Command{
	Use:   "status [REMOTE]",
	Short: "Query the status of a remote instance",
	Long: `Query the remote VPS for connectivity, daemon
behaviour, and other information.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, err := GetProjectConfigFromDisk()
		if err != nil {
			log.WithError(err)
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
	remoteCmd.AddCommand(statusCmd)

	homeEnvVar := os.Getenv("HOME")
	sshDir := filepath.Join(homeEnvVar, ".ssh")
	defaultSSHLoc := filepath.Join(sshDir, "id_rsa")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	remoteCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	addCmd.Flags().StringP("user", "u", "root", "User for SSH access")
	addCmd.Flags().StringP("identity", "i", defaultSSHLoc, "PEM file location")
	addCmd.Flags().StringP("port", "p", defaultDaemonPort, "Daemon port")
}

// Bootstrap configures a remote vps for continuous deployment
// by installing docker, starting the daemon and building a
// public-private key-pair. It outputs configuration information
// for the user.
func (remote *RemoteVPS) Bootstrap(runner SSHSession, name string, config *Config) error {
	println("Bootstrapping remote " + name)

	// Generate a session for each command.
	println("Installing docker")
	err := remote.InstallDocker(runner)
	if err != nil {
		return err
	}

	println("Starting daemon")
	if err != nil {
		return err
	}
	err = remote.DaemonUp(runner, remote.Port)
	if err != nil {
		return err
	}

	println("Building deploy key\n")
	if err != nil {
		return err
	}
	pub, err := remote.KeyGen(runner)
	if err != nil {
		return err
	}

	println("Fetching daemon API token")
	token, err := remote.GetDaemonAPIToken(runner)
	if err != nil {
		return err
	}

	config.DaemonAPIToken = token
	f, err := GetConfigFile()
	defer f.Close()
	_, err = config.Write(f)
	if err != nil {
		return err
	}

	println("Daemon running on instance")

	// Output deploy key to user.
	println("GitHub Deploy Key (add here https://www.github.com/<your_repo>/settings/keys/new): ")
	println(pub.String())

	// Output Webhook url to user.
	println("GitHub WebHook URL (add here https://www.github.com/<your_repo>/settings/hooks/new): ")
	println("http://" + remote.IP + ":" + remote.Port)
	println("Github WebHook Secret: " + defaultSecret + "\n")

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
func (remote *RemoteVPS) RunSSHCommand(runner SSHSession, remoteCmd string) (
	*bytes.Buffer, *bytes.Buffer, error) {
	return runner.Run(remoteCmd)
}

// InstallDocker installs docker on a remote vps.
func (remote *RemoteVPS) InstallDocker(session SSHSession) error {
	// Collect assets (docker shell script)
	installDockerSh, err := Asset("cmd/bootstrap/docker.sh")
	if err != nil {
		return err
	}

	// Install docker.
	cmdStr := string(installDockerSh)
	_, stderr, err := remote.RunSSHCommand(session, cmdStr)
	if err != nil {
		println(stderr)
		return err
	}

	return nil
}

// DaemonUp brings the daemon up on the remote instance.
func (remote *RemoteVPS) DaemonUp(session SSHSession, daemonPort string) error {
	// Collect assets (deamon-up shell script)
	daemonCmd, err := Asset("cmd/bootstrap/daemon-up.sh")
	if err != nil {
		return err
	}

	// Run inertia daemon.
	daemonCmdStr := fmt.Sprintf(string(daemonCmd), daemonPort)
	_, stderr, err := remote.RunSSHCommand(session, daemonCmdStr)
	if err != nil {
		println(stderr)
		return err
	}

	return nil
}

// KeyGen creates a public-private key-pair on the remote vps
// and returns the public key.
func (remote *RemoteVPS) KeyGen(session SSHSession) (*bytes.Buffer, error) {
	// Collect assets (keygen shell script)
	keygenSh, err := Asset("cmd/bootstrap/keygen.sh")
	if err != nil {
		return nil, err
	}

	// Create deploy key.
	result, stderr, err := remote.RunSSHCommand(session, string(keygenSh))

	if err != nil {
		log.Println(stderr)
		return nil, err
	}

	return result, nil
}

// DaemonDown brings the daemon down on the remote instance
func (remote *RemoteVPS) DaemonDown(session SSHSession) error {
	// Collect assets (deamon-up shell script)
	daemonCmd, err := Asset("cmd/bootstrap/daemon-down.sh")
	if err != nil {
		return err
	}

	_, stderr, err := remote.RunSSHCommand(session, string(daemonCmd))
	if err != nil {
		log.Println(stderr)
		return err
	}

	return nil
}

// GetDaemonAPIToken returns the daemon API token for RESTful access
// to the daemon.
func (remote *RemoteVPS) GetDaemonAPIToken(session SSHSession) (string, error) {
	// Collect asset (token.sh script)
	daemonCmd, err := Asset("cmd/bootstrap/token.sh")
	if err != nil {
		return "", err
	}

	stdout, stderr, err := remote.RunSSHCommand(session, string(daemonCmd))
	if err != nil {
		log.Println(stderr)
		return "", err
	}

	// There may be a newline, remove it.
	return strings.TrimSuffix(stdout.String(), "\n"), nil
}

// addNewRemote adds a new remote to the project config file.
func addNewRemote(name, IP, user, pemLoc, port string) error {
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
		Port: port,
	}

	f, err := GetConfigFile()
	defer f.Close()
	_, err = config.Write(f)
	if err != nil {
		return err
	}

	println("Remote '" + name + "' added.")

	return nil
}

// Stubbed out for testing.
func getSSHSession(PEM, IP, user string) (*ssh.Session, error) {
	privateKey, err := ioutil.ReadFile(PEM)
	if err != nil {
		return nil, err
	}

	cfg, err := getSSHConfig(privateKey, user)
	client, err := ssh.Dial("tcp", IP+":22", cfg)
	if err != nil {
		return nil, err
	}

	// Create a session. It is one session per command.
	return client.NewSession()
}

// getSSHConfig returns SSH configuration for the remote.
func getSSHConfig(privateKey []byte, user string) (*ssh.ClientConfig, error) {
	key, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	// Authentication
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		// TODO: We need to replace this with a callback
		// to verify the host key. A security vulnerability
		// comes from the fact that we recieve a public key
		// from the server and we add it to our GitHub.
		// This gives the server readonly access to our
		// GitHub account. We need to know who we're
		// connecting to.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
