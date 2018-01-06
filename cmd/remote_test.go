package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

// SSHRunner runs commands over SSH and captures results.
type mockSSHRunner struct {
	r        *RemoteVPS
	LastCall string
}

// Run runs a command remotely.
func (runner *mockSSHRunner) Run(cmd string) (*bytes.Buffer, *bytes.Buffer, error) {
	runner.LastCall = cmd
	return nil, nil, nil
}

func TestRunSSHCommand(t *testing.T) {
	remote := getTestRemote()
	session := mockSSHRunner{r: remote}
	cmd := "ls -lsa"
	_, _, err := remote.RunSSHCommand(&session, cmd)

	assert.Nil(t, err)
	assert.Equal(t, session.LastCall, cmd)
}

func TestInstallDocker(t *testing.T) {
	remote := getTestRemote()
	script, err := ioutil.ReadFile("bootstrap/docker.sh")
	assert.Nil(t, err)

	// Make sure the right command is run.
	session := mockSSHRunner{r: remote}
	remote.InstallDocker(&session)
	assert.Equal(t, session.LastCall, string(script))
}

func TestDaemonDown(t *testing.T) {
	remote := getTestRemote()
	script, err := ioutil.ReadFile("bootstrap/daemon-up.sh")
	assert.Nil(t, err)

	actualCommand := fmt.Sprintf(string(script), "8081")

	// Make sure the right command is run.
	session := mockSSHRunner{r: remote}

	// Make sure the right command is run.
	err = remote.DaemonUp(&session, "8081")
	assert.Nil(t, err)
	assert.Equal(t, session.LastCall, actualCommand)
}

func TestKeyGen(t *testing.T) {
	remote := getTestRemote()
	script, err := ioutil.ReadFile("bootstrap/daemon-down.sh")
	assert.Nil(t, err)

	// Make sure the right command is run.
	session := mockSSHRunner{r: remote}

	// Make sure the right command is run.
	err = remote.DaemonDown(&session)
	assert.Nil(t, err)
	assert.Equal(t, session.LastCall, string(script))
}

/*
func TestBootstrap(t *testing.T) {
	remote := getTestRemote()
	script, err := ioutil.ReadFile("bootstrap/token.sh")
	assert.Nil(t, err)

	// Make sure the right command is run.
	session := mockSSHRunner{r: remote}
	err = remote.Bootstrap(&session, "gcloud", &Config{})
	assert.Nil(t, err)

	// Just check last call.
	assert.Nil(t, err)
	assert.Equal(t, session.LastCall, string(script))
}
*/
