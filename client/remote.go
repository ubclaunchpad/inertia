package client

import (
	"bytes"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// RemoteVPS contains parameters for the VPS
type RemoteVPS struct {
	Name             string        `toml:"name"`
	IP               string        `toml:"IP"`
	User             string        `toml:"user"`
	PEM              string        `toml:"pemfile"`
	Branch           string        `toml:"branch"`
	SSHPort          string        `toml:"ssh_port"`
	ProjectDirectory string        `toml:"project_directory"`
	Daemon           *DaemonConfig `toml:"daemon"`
}

// DaemonConfig contains parameters for the Daemon
type DaemonConfig struct {
	Port   string `toml:"port"`
	Token  string `toml:"token"`
	Secret string `toml:"secret"`
}

// GetHost creates the user@IP string.
func (remote *RemoteVPS) GetHost() string {
	return remote.User + "@" + remote.IP
}

// GetIPAndPort creates the IP:Port string.
func (remote *RemoteVPS) GetIPAndPort() string {
	return remote.IP + ":" + remote.Daemon.Port
}

// installDocker installs docker on a remote vps.
func (remote *RemoteVPS) installDocker(session SSHSession) error {
	installDockerSh, err := Asset("client/bootstrap/docker.sh")
	if err != nil {
		return err
	}

	// Install docker.
	cmdStr := string(installDockerSh)
	_, stderr, err := session.Run(cmdStr)
	if err != nil {
		println(stderr.String())
		return err
	}

	return nil
}

// keyGen creates a public-private key-pair on the remote vps
// and returns the public key.
func (remote *RemoteVPS) keyGen(session SSHSession) (*bytes.Buffer, error) {
	scriptBytes, err := Asset("client/bootstrap/keygen.sh")
	if err != nil {
		return nil, err
	}

	// Create deploy key.
	result, stderr, err := session.Run(string(scriptBytes))

	if err != nil {
		log.Println(stderr.String())
		return nil, err
	}

	return result, nil
}

// getDaemonAPIToken returns the daemon API token for RESTful access
// to the daemon.
func (remote *RemoteVPS) getDaemonAPIToken(session SSHSession, daemonVersion string) (string, error) {
	scriptBytes, err := Asset("client/bootstrap/token.sh")
	if err != nil {
		return "", err
	}
	daemonCmdStr := fmt.Sprintf(string(scriptBytes), daemonVersion)

	stdout, stderr, err := session.Run(daemonCmdStr)
	if err != nil {
		log.Println(stderr.String())
		return "", err
	}

	// There may be a newline, remove it.
	return strings.TrimSuffix(stdout.String(), "\n"), nil
}
