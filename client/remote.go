package client

import (
	"bytes"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/ubclaunchpad/inertia/common"
)

// RemoteVPS contains parameters for the VPS
type RemoteVPS struct {
	Name   string        `toml:"name"`
	IP     string        `toml:"IP"`
	User   string        `toml:"user"`
	PEM    string        `toml:"pemfile"`
	Branch string        `toml:"branch"`
	Daemon *DaemonConfig `toml:"daemon"`
}

// DaemonConfig contains parameters for the Daemon
type DaemonConfig struct {
	Port    string `toml:"port"`
	SSHPort string `toml:"ssh_port"`
	Token   string `toml:"token"`
}

// GetHost creates the user@IP string.
func (remote *RemoteVPS) GetHost() string {
	return remote.User + "@" + remote.IP
}

// GetIPAndPort creates the IP:Port string.
func (remote *RemoteVPS) GetIPAndPort() string {
	return remote.IP + ":" + remote.Daemon.Port
}

// RunSSHCommand runs a command remotely.
func (remote *RemoteVPS) RunSSHCommand(runner SSHSession, remoteCmd string) (
	*bytes.Buffer, *bytes.Buffer, error,
) {
	return runner.Run(remoteCmd)
}

// Bootstrap configures a remote vps for continuous deployment
// by installing docker, starting the daemon and building a
// public-private key-pair. It outputs configuration information
// for the user.
func (remote *RemoteVPS) Bootstrap(runner SSHSession, name string, config *Config) error {
	println("Setting up remote " + name)

	// Generate a session for each command.
	println("Installing docker...")
	err := remote.installDocker(runner)
	if err != nil {
		return err
	}

	println("Building deploy key...")
	if err != nil {
		return err
	}
	pub, err := remote.keyGen(runner)
	if err != nil {
		return err
	}

	println("Fetching daemon API token...")
	token, err := remote.getDaemonAPIToken(runner, config.Version)
	if err != nil {
		return err
	}

	remote.Daemon.Token = token
	err = config.Write()
	if err != nil {
		return err
	}

	println("Setting up SSL certificate...")
	// TODO

	println("Starting daemon...")
	if err != nil {
		return err
	}
	err = remote.DaemonUp(runner, config.Version, remote.Daemon.Port)
	if err != nil {
		return err
	}

	println("Daemon running on instance!")

	// Output deploy key to user.
	println("GitHub Deploy Key (add here https://www.github.com/<your_repo>/settings/keys/new): ")
	println(pub.String())

	// Output Webhook url to user.
	println("GitHub WebHook URL (add here https://www.github.com/<your_repo>/settings/hooks/new): ")
	println("http://" + remote.IP + ":" + remote.Daemon.Port)
	println("Github WebHook Secret: " + common.DefaultSecret + "\n")

	println("Inertia daemon successfully deployed! Add your webhook url and deploy\nkey to enable continuous deployment.")
	fmt.Printf("Then run 'inertia %s up' to deploy your application.\n", name)

	return nil
}

// DaemonUp brings the daemon up on the remote instance.
func (remote *RemoteVPS) DaemonUp(session SSHSession, daemonVersion, daemonPort string) error {
	daemonCmd, err := Asset("client/bootstrap/daemon-up.sh")
	if err != nil {
		return err
	}

	// Run inertia daemon.
	daemonCmdStr := fmt.Sprintf(string(daemonCmd), daemonVersion, daemonPort)
	_, stderr, err := remote.RunSSHCommand(session, daemonCmdStr)
	if err != nil {
		println(stderr.String())
		return err
	}

	return nil
}

// DaemonDown brings the daemon down on the remote instance
func (remote *RemoteVPS) DaemonDown(session SSHSession) error {
	daemonCmd, err := Asset("client/bootstrap/daemon-down.sh")
	if err != nil {
		return err
	}

	_, stderr, err := remote.RunSSHCommand(session, string(daemonCmd))
	if err != nil {
		println(stderr.String())
		return err
	}

	return nil
}

// installDocker installs docker on a remote vps.
func (remote *RemoteVPS) installDocker(session SSHSession) error {
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

// keyGen creates a public-private key-pair on the remote vps
// and returns the public key.
func (remote *RemoteVPS) keyGen(session SSHSession) (*bytes.Buffer, error) {
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

// getDaemonAPIToken returns the daemon API token for RESTful access
// to the daemon.
func (remote *RemoteVPS) getDaemonAPIToken(session SSHSession, daemonVersion string) (string, error) {
	// Collect asset (token.sh script)
	daemonCmd, err := Asset("client/bootstrap/token.sh")
	if err != nil {
		return "", err
	}
	daemonCmdStr := fmt.Sprintf(string(daemonCmd), daemonVersion)

	stdout, stderr, err := remote.RunSSHCommand(session, daemonCmdStr)
	if err != nil {
		log.Println(stderr.String())
		return "", err
	}

	// There may be a newline, remove it.
	return strings.TrimSuffix(stdout.String(), "\n"), nil
}
