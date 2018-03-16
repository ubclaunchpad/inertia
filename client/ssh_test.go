package client

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// SSHRunner runs commands over SSH and captures results.
type mockSSHRunner struct {
	r     *RemoteVPS
	Calls []string
}

// Run runs a command remotely.
func (runner *mockSSHRunner) Run(cmd string) (*bytes.Buffer, *bytes.Buffer, error) {
	runner.Calls = append(runner.Calls, cmd)
	return nil, nil, nil
}

func TestRunSSHCommand(t *testing.T) {
	remote := getInstrumentedTestRemote()
	session := mockSSHRunner{r: remote}
	cmd := "ls -lsa"
	_, _, err := remote.RunSSHCommand(&session, cmd)

	assert.Nil(t, err)
	assert.Equal(t, cmd, session.Calls[0])
}
