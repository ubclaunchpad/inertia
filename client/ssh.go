package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
)

// SSHSession can run remote commands over SSH
type SSHSession interface {
	Run(cmd string) (*bytes.Buffer, *bytes.Buffer, error)
	RunStream(cmd string, interactive bool) error
	RunSession() error
	CopyFile(r io.Reader, remotePath string, permissions string, size int64) error
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
	session, err := getSSHSession(runner.r.PEM, runner.r.IP, runner.r.Daemon.SSHPort, runner.r.User)
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
	session, err := getSSHSession(runner.r.PEM, runner.r.IP, runner.r.Daemon.SSHPort, runner.r.User)
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
	session, err := getSSHSession(runner.r.PEM, runner.r.IP, runner.r.Daemon.SSHPort, runner.r.User)
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
func (runner *SSHRunner) CopyFile(fileReader io.Reader, remotePath string, permissions string, size int64) error {
	contentsBytes, _ := ioutil.ReadAll(fileReader)
	reader := bytes.NewReader(contentsBytes)

	session, err := getSSHSession(runner.r.PEM, runner.r.IP, runner.r.Daemon.SSHPort, runner.r.User)
	if err != nil {
		return err
	}

	filename := path.Base(remotePath)
	directory := path.Dir(remotePath)

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintln(w, "C"+permissions, size, filename)
		io.Copy(w, reader)
		fmt.Fprintln(w, "\x00")
	}()

	session.Run("/usr/bin/scp -t " + directory)
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
