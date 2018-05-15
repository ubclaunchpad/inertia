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
func (remote *RemoteVPS) Bootstrap(runner SSHSession, repoName string, config *Config) error {
	println("Setting up remote \"" + remote.Name + "\" at " + remote.IP)

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

	println("\nInertia has been set up and daemon is running on remote!")
	println("You may have to wait briefly for Inertia to set up some dependencies.")
	fmt.Printf("Use 'inertia %s logs --stream' to check on the daemon's setup progress.\n\n", remote.Name)

	println("=============================\n")

	// Output deploy key to user.
	println(">> GitHub Deploy Key (add to https://www.github.com/" + repoName + "/settings/keys/new): ")
	println(pub.String())

	// Output Webhook url to user.
	println(">> GitHub WebHook URL (add to https://www.github.com/" + repoName + "/settings/hooks/new): ")
	println("WebHook Address:  https://" + remote.IP + ":" + remote.Daemon.Port + "/webhook")
	println("WebHook Secret:   " + remote.Daemon.Secret)
	println(`Note that you will have to disable SSH verification in your webhook
settings - Inertia uses self-signed certificates that GitHub won't
be able to verify.` + "\n")

	println(`Inertia daemon successfully deployed! Add your webhook url and deploy
key to enable continuous deployment.`)
	fmt.Printf("Then run 'inertia %s up' to deploy your application.\n", remote.Name)

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
