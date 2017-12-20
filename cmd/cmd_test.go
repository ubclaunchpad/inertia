package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestRemote() *RemoteVPS {
	return &RemoteVPS{
		IP:   "127.0.0.1",
		PEM:  "/Users/me/and/my/pem",
		User: "me",
		Port: "5555",
	}
}

func TestConfigWrite(t *testing.T) {
	config := &Config{
		CurrentRemoteName: "test",
		CurrentRemoteVPS:  getTestRemote(),
	}

	var f bytes.Buffer
	n, err := config.Write(&f)

	assert.Nil(t, err)
	assert.Equal(t, n, 98)
}

// This disgusting closure hijacks the exec.Command call, and stuffs
// some of the call args into my buffer.
// https://npf.io/2015/06/testing-exec-command/
func fakeCommand(buf *bytes.Buffer) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)

		// Capture the command piped in (last arg).
		buf.WriteString(cs[len(cs)-1])
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
}

func TestRunSSHCommand(t *testing.T) {
	remote := getTestRemote()

	var capturedCommand bytes.Buffer
	execCommand = fakeCommand(&capturedCommand)
	defer func() { execCommand = exec.Command }()

	cmd := "ls -lsa"

	_, _, err := remote.RunSSHCommand(cmd)
	if err != nil {
		t.Errorf("Expected nil error, got %#v", err)
	}

	assert.Equal(t, capturedCommand.String(), cmd)
}

func TestInstallDocker(t *testing.T) {
	remote := getTestRemote()

	var capturedCommand, actualCommand bytes.Buffer
	execCommand = fakeCommand(&capturedCommand)
	defer func() { execCommand = exec.Command }()

	script, err := os.Open("bootstrap/docker.sh")
	assert.Nil(t, err)
	defer script.Close()

	_, err = io.Copy(&actualCommand, script)
	assert.Nil(t, err)

	// Make sure the right command is run.
	remote.InstallDocker()
	assert.Equal(t, actualCommand, capturedCommand)
}

func TestDaemonDown(t *testing.T) {
	remote := getTestRemote()

	var capturedCommand, actualCommand bytes.Buffer
	execCommand = fakeCommand(&capturedCommand)
	defer func() { execCommand = exec.Command }()

	script, err := os.Open("bootstrap/daemon-down.sh")
	assert.Nil(t, err)
	defer script.Close()

	_, err = io.Copy(&actualCommand, script)
	assert.Nil(t, err)

	// Make sure the right command is run.
	remote.DaemonDown()
	assert.Equal(t, actualCommand, capturedCommand)
}

func TestKeyGen(t *testing.T) {
	remote := getTestRemote()

	var capturedCommand, actualCommand bytes.Buffer
	execCommand = fakeCommand(&capturedCommand)
	defer func() { execCommand = exec.Command }()

	script, err := os.Open("bootstrap/keygen.sh")
	assert.Nil(t, err)
	defer script.Close()

	_, err = io.Copy(&actualCommand, script)
	assert.Nil(t, err)

	// Make sure the right command is run.
	remote.KeyGen()
	assert.Equal(t, actualCommand, capturedCommand)
}
