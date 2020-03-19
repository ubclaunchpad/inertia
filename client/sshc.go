package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ubclaunchpad/inertia/cfg"
	internal "github.com/ubclaunchpad/inertia/client/internal"
	"github.com/ubclaunchpad/inertia/client/runner"
)

// SSHClient implements Inertia's SSH commands
type SSHClient struct {
	om    sync.Mutex
	out   io.Writer
	debug bool

	remote *cfg.Remote
	ssh    runner.SSHSession
}

// GetRunner returns the SSH client's underlying session
func (s *SSHClient) GetRunner() runner.SSHSession { return s.ssh }

// DaemonUp brings the daemon up on the remote instance.
func (s *SSHClient) DaemonUp() error {
	scriptBytes, err := internal.ReadFile("client/scripts/daemon-up.sh")
	if err != nil {
		return fmt.Errorf("could not initialize script: %w", err)
	}
	var daemonCmdStr = fmt.Sprintf(string(scriptBytes),
		s.remote.Version, s.remote.Daemon.Port, s.remote.IP, s.remote.Daemon.WebHookSecret)
	return s.ssh.RunStream(daemonCmdStr, false)
}

// DaemonDown brings the daemon down on the remote instance
func (s *SSHClient) DaemonDown() error {
	scriptBytes, err := internal.ReadFile("client/scripts/daemon-down.sh")
	if err != nil {
		return fmt.Errorf("could not initialize script: %w", err)
	}

	stdout, stderr, err := s.ssh.Run(string(scriptBytes))
	s.debugStdout("daemon-down.sh", stdout)
	s.debugStderr("daemon-down.sh", stderr)
	if err != nil {
		return fmt.Errorf("daemon shutdown failed: %s", err.Error())
	}

	return nil
}

// InstallDocker installs docker on a remote vps.
func (s *SSHClient) InstallDocker() error {
	installDockerSh, err := internal.ReadFile("client/scripts/docker.sh")
	if err != nil {
		return fmt.Errorf("could not initialize script: %w", err)
	}

	// Install docker.
	cmdStr := string(installDockerSh)
	if err = s.ssh.RunStream(cmdStr, false); err != nil {
		return fmt.Errorf("docker installation failed: %w", err)
	}

	return nil
}

// GenerateKeys creates a public-private key-pair on the remote vps and returns
// the public key.
func (s *SSHClient) GenerateKeys() (string, error) {
	if s.ssh == nil {
		return "", errors.New("client not configured for SSH access")
	}

	scriptBytes, err := internal.ReadFile("client/scripts/keygen.sh")
	if err != nil {
		return "", fmt.Errorf("could not initialize script: %w", err)
	}

	// Create deploy key.
	stdout, stderr, err := s.ssh.Run(string(scriptBytes))
	s.debugStdout("keygen.sh", stdout)
	s.debugStderr("keygen.sh", stderr)
	if err != nil {
		return "", fmt.Errorf("key generation failed: %w", err)
	}

	return stdout.String(), nil
}

// AssignAPIToken generates an API token and assigns it to client.Remote
func (s *SSHClient) AssignAPIToken() error {
	scriptBytes, err := internal.ReadFile("client/scripts/token.sh")
	if err != nil {
		return fmt.Errorf("could not initialize script: %w", err)
	}
	daemonCmdStr := fmt.Sprintf(string(scriptBytes), s.remote.Version)
	stdout, stderr, err := s.ssh.Run(daemonCmdStr)
	s.debugStdout("token.sh", stdout)
	s.debugStderr("token.sh", stderr)
	if err != nil {
		return fmt.Errorf("api token generation failed: %w", err)
	}

	// There may be a newline, remove it.
	s.remote.Daemon.Token = strings.TrimSuffix(stdout.String(), "\n")
	return nil
}

// UninstallInertia removes the inertia/ directory on the remote instance
func (s *SSHClient) UninstallInertia() error {
	scriptBytes, err := internal.ReadFile("client/scripts/inertia-down.sh")
	if err != nil {
		return fmt.Errorf("could not initialize script: %w", err)
	}

	stdout, stderr, err := s.ssh.Run(string(scriptBytes))
	s.debugStdout("inertia-down.sh", stdout)
	s.debugStderr("inertia-down.sh", stderr)
	if err != nil {
		return fmt.Errorf("inertia shutdown failed: %w", err)
	}

	return nil
}

// debugf logs to the client's output if debug is enabled
func (s *SSHClient) debugf(format string, args ...interface{}) {
	if s.debug {
		s.om.Lock()
		fmt.Fprintf(s.out, "DEBUG: "+format+"\n", args...)
		s.om.Unlock()
	}
}

func (s *SSHClient) debugStderr(script string, out *bytes.Buffer) {
	if out != nil && out.Len() > 0 {
		s.debugf("%s stderr:\n>>>\n%s\n<<<", script, out.String())
	}
}

func (s *SSHClient) debugStdout(script string, out *bytes.Buffer) {
	if out != nil && out.Len() > 0 {
		s.debugf("%s stdout:\n>>>\n%s\n<<<", script, out.String())
	}
}
