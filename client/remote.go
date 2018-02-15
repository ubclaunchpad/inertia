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
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/ubclaunchpad/inertia/common"
	"golang.org/x/crypto/ssh"
)

// RemoteVPS contains parameters for the VPS
type RemoteVPS struct {
	User       string
	IP         string
	SSHPort    string
	PEM        string
	DaemonPort string
}

// SSHSession can run remote commands over SSH
type SSHSession interface {
	Run(cmd string) (*bytes.Buffer, *bytes.Buffer, error)
}

// SSHRunner runs commands over SSH and captures results.
type SSHRunner struct {
	r *RemoteVPS
}

// NewSSHRunner returns a new SSHRunner
func NewSSHRunner(r *RemoteVPS) *SSHRunner {
	return &SSHRunner{r: r}
}

// Run runs a command remotely.
func (runner *SSHRunner) Run(cmd string) (*bytes.Buffer, *bytes.Buffer, error) {
	session, err := getSSHSession(runner.r.PEM, runner.r.IP, runner.r.SSHPort, runner.r.User)
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
	err = remote.DaemonUp(runner, remote.DaemonPort)
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
	_, err = config.Write()
	if err != nil {
		return err
	}

	println("Daemon running on instance")

	// Output deploy key to user.
	println("GitHub Deploy Key (add here https://www.github.com/<your_repo>/settings/keys/new): ")
	println(pub.String())

	// Output Webhook url to user.
	println("GitHub WebHook URL (add here https://www.github.com/<your_repo>/settings/hooks/new): ")
	println("http://" + remote.IP + ":" + remote.DaemonPort)
	println("Github WebHook Secret: " + common.DefaultSecret + "\n")

	println("Inertia daemon successfully deployed! Add your webhook url and deploy\nkey to enable continuous deployment.")
	fmt.Printf("Then run 'inertia %s up' to deploy your application.\n", name)

	return nil
}

// GetHost creates the user@IP string.
func (remote *RemoteVPS) GetHost() string {
	return remote.User + "@" + remote.IP
}

// GetIPAndPort creates the IP:Port string.
func (remote *RemoteVPS) GetIPAndPort() string {
	return remote.IP + ":" + remote.DaemonPort
}

// RunSSHCommand runs a command remotely.
func (remote *RemoteVPS) RunSSHCommand(runner SSHSession, remoteCmd string) (
	*bytes.Buffer, *bytes.Buffer, error) {
	return runner.Run(remoteCmd)
}

// InstallDocker installs docker on a remote vps.
func (remote *RemoteVPS) InstallDocker(session SSHSession) error {
	// Collect assets (docker shell script)
	installDockerSh, err := Asset("client/bootstrap/docker.sh")
	if err != nil {
		return err
	}

	// Install docker.
	cmdStr := string(installDockerSh)
	_, stderr, err := remote.RunSSHCommand(session, cmdStr)
	if err != nil {
		println(stderr.String())
		return err
	}

	return nil
}

// DaemonUp brings the daemon up on the remote instance.
func (remote *RemoteVPS) DaemonUp(session SSHSession, daemonPort string) error {
	// Collect assets (deamon-up shell script)
	daemonCmd, err := Asset("client/bootstrap/daemon-up.sh")
	if err != nil {
		return err
	}

	// Run inertia daemon.
	daemonCmdStr := fmt.Sprintf(string(daemonCmd), daemonPort)
	_, stderr, err := remote.RunSSHCommand(session, daemonCmdStr)
	if err != nil {
		println(stderr.String())
		return err
	}

	return nil
}

// KeyGen creates a public-private key-pair on the remote vps
// and returns the public key.
func (remote *RemoteVPS) KeyGen(session SSHSession) (*bytes.Buffer, error) {
	// Collect assets (keygen shell script)
	keygenSh, err := Asset("client/bootstrap/keygen.sh")
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
	daemonCmd, err := Asset("client/bootstrap/daemon-down.sh")
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
	daemonCmd, err := Asset("client/bootstrap/token.sh")
	if err != nil {
		return "", err
	}

	stdout, stderr, err := remote.RunSSHCommand(session, string(daemonCmd))
	if err != nil {
		log.Println(stderr.String())
		return "", err
	}

	// There may be a newline, remove it.
	return strings.TrimSuffix(stdout.String(), "\n"), nil
}

// AddNewRemote adds a new remote to the project config file.
func AddNewRemote(name, IP, sshPort, user, pemLoc, port string) error {
	// Just wipe configuration for MVP.
	config, err := GetProjectConfigFromDisk()
	if err != nil {
		return err
	}

	config.CurrentRemoteName = name
	config.CurrentRemoteVPS = &RemoteVPS{
		IP:         IP,
		SSHPort:    sshPort,
		User:       user,
		PEM:        pemLoc,
		DaemonPort: port,
	}

	_, err = config.Write()
	return err
}

// Stubbed out for testing.
func getSSHSession(PEM, IP, sshPort, user string) (*ssh.Session, error) {
	privateKey, err := ioutil.ReadFile(PEM)
	if err != nil {
		return nil, err
	}

	cfg, err := getSSHConfig(privateKey, user)
	if err != nil {
		return nil, err
	}

	client, err := ssh.Dial("tcp", IP+":"+sshPort, cfg)
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
		// comes from the fact that we receive a public key
		// from the server and we add it to our GitHub.
		// This gives the server readonly access to our
		// GitHub account. We need to know who we're
		// connecting to.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
