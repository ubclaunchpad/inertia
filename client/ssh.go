package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ubclaunchpad/inertia/cfg"
	"golang.org/x/crypto/ssh"
)

// SSHSession can run remote commands over SSH
type SSHSession interface {
	Run(cmd string) (*bytes.Buffer, *bytes.Buffer, error)
	RunStream(cmd string, interactive bool) error
	RunSession() error
	CopyFile(f io.Reader, remotePath string, permissions string) error
}

// SSHRunner runs commands over SSH and captures results.
type SSHRunner struct {
	pem     string
	user    string
	ip      string
	sshPort string
}

// NewSSHRunner returns a new SSHRunner
func NewSSHRunner(r *cfg.RemoteVPS) *SSHRunner {
	if r != nil {
		return &SSHRunner{
			pem:     r.PEM,
			user:    r.User,
			ip:      r.IP,
			sshPort: r.SSHPort,
		}
	}
	return &SSHRunner{}
}

// Run runs a command remotely.
func (runner *SSHRunner) Run(cmd string) (cmdout *bytes.Buffer, cmderr *bytes.Buffer, err error) {
	session, err := getSSHSession(runner.pem, runner.ip, runner.sshPort, runner.user)
	if err != nil {
		return nil, nil, err
	}

	// Capture result.
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Execute command.
	err = session.Run(cmd)
	return &stdout, &stderr, err
}

// RunStream remotely executes given command, streaming its output
// and opening up an optionally interactive session
func (runner *SSHRunner) RunStream(cmd string, interactive bool) error {
	session, err := getSSHSession(runner.pem, runner.ip, runner.sshPort, runner.user)
	if err != nil {
		return err
	}

	// Pipe input and outputs.
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	if interactive {
		session.Stdin = os.Stdin
	}

	// Execute command.
	return session.Run(cmd)
}

// RunSession sets up a SSH shell to the remote
func (runner *SSHRunner) RunSession() error {
	session, err := getSSHSession(runner.pem, runner.ip, runner.sshPort, runner.user)
	if err != nil {
		return err
	}

	// Set IO
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	in, _ := session.StdinPipe()

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return err
	}

	// Start remote shell
	if err := session.Shell(); err != nil {
		return err
	}

	// Accepting commands
	for {
		reader := bufio.NewReader(os.Stdin)
		str, _ := reader.ReadString('\n')
		fmt.Fprint(in, str)
	}
}

// CopyFile copies given reader to remote
func (runner *SSHRunner) CopyFile(file io.Reader, remotePath string, permissions string) error {
	// Open and read file
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(contents)

	// Set up
	filename := filepath.Base(remotePath)
	directory := filepath.Dir(remotePath)
	session, err := getSSHSession(runner.pem, runner.ip, runner.sshPort, runner.user)
	if err != nil {
		return err
	}

	// Send file contents
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintln(w, "C"+permissions, len(contents), filename)
		io.Copy(w, reader)
		fmt.Fprintln(w, "\x00")
	}()
	session.Run("mkdir -p " + directory + "; /usr/bin/scp -t " + directory)
	return nil
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
