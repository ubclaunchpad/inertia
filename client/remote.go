package client

import (
	"bytes"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
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
	Secret  string `toml:"secret"`
}

// GetHost creates the user@IP string.
func (remote *RemoteVPS) GetHost() string {
	return remote.User + "@" + remote.IP
}

// GetIPAndPort creates the IP:Port string.
func (remote *RemoteVPS) GetIPAndPort() string {
	return remote.IP + ":" + remote.Daemon.Port
}

// Bootstrap configures a remote vps for continuous deployment
// by installing docker, starting the daemon and building a
// public-private key-pair. It outputs configuration information
// for the user.
func (remote *RemoteVPS) Bootstrap(runner SSHSession, name string, config *Config) error {
	println("Setting up remote \"" + name + "\" at " + remote.IP)

	println(">> Step 1/4: Installing docker...")
	err := remote.installDocker(runner)
	if err != nil {
		return err
	}

	println("\n>> Step 2/4: Building deploy key...")
	if err != nil {
		return err
	}
	pub, err := remote.keyGen(runner)
	if err != nil {
		return err
	}

	// This step needs to run before any other commands that rely on
	// the daemon image, since the daemon is loaded here.
	println("\n>> Step 3/4: Starting daemon...")
	if err != nil {
		return err
	}
	err = remote.DaemonUp(runner, config.Version, remote.IP, remote.Daemon.Port)
	if err != nil {
		return err
	}

	println("\n>> Step 4/4: Fetching daemon API token...")
	token, err := remote.getDaemonAPIToken(runner, config.Version)
	if err != nil {
		return err
	}
	remote.Daemon.Token = token
	err = config.Write()
	if err != nil {
		return err
	}

	println("\nInertia is now set up and the daemon is running on your remote!\n")

	println("=============================\n")

	// Output deploy key to user.
	println("GitHub Deploy Key (add to https://www.github.com/<your_repo>/settings/keys/new): ")
	println(pub.String())

	// Output webhook url to user.
	println("GitHub WebHook URL (add to https://www.github.com/<your_repo>/settings/hooks/new): ")
	println("http://" + remote.IP + ":" + remote.Daemon.Port)
	println("Github WebHook Secret: " + remote.Daemon.Secret + "\n")

	println("Inertia daemon successfully deployed! Add your webhook url and deploy\nkey to enable continuous deployment.")
	fmt.Printf("Then run 'inertia %s up' to deploy your application!\n", name)

	return nil
}

// DaemonUp brings the daemon up on the remote instance.
func (remote *RemoteVPS) DaemonUp(session SSHSession, daemonVersion, host, daemonPort string) error {
	scriptBytes, err := Asset("client/bootstrap/daemon-up.sh")
	if err != nil {
		return err
	}

	// Run inertia daemon.
	daemonCmdStr := fmt.Sprintf(string(scriptBytes), daemonVersion, daemonPort, host)
	return session.RunStream(daemonCmdStr, false)
}

// DaemonDown brings the daemon down on the remote instance
func (remote *RemoteVPS) DaemonDown(session SSHSession) error {
	scriptBytes, err := Asset("client/bootstrap/daemon-down.sh")
	if err != nil {
		return err
	}

	_, stderr, err := session.Run(string(scriptBytes))
	if err != nil {
		println(stderr.String())
		return err
	}

	return nil
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
