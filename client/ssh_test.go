package client

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockSSHRunner is a mocked out implementation of SSHSession
type mockSSHRunner struct {
	r     *RemoteVPS
	Calls []string
}

func (runner *mockSSHRunner) Run(cmd string) (*bytes.Buffer, *bytes.Buffer, error) {
	runner.Calls = append(runner.Calls, cmd)
	return nil, nil, nil
}

func (runner *mockSSHRunner) RunInteractive(cmd string) error {
	runner.Calls = append(runner.Calls, cmd)
	return nil
}

func (runner *mockSSHRunner) RunSession() error {
	return nil
}

func TestRun(t *testing.T) {
	remote := getInstrumentedTestRemote()
	session := mockSSHRunner{r: remote}
	cmd := "ls -lsa"

	_, _, err := session.Run(cmd)
	assert.Nil(t, err)
	assert.Equal(t, cmd, session.Calls[0])
}

func TestRunInteractive(t *testing.T) {
	remote := getInstrumentedTestRemote()
	session := mockSSHRunner{r: remote}
	cmd := "ls -lsa"

	err := session.RunInteractive(cmd)
	assert.Nil(t, err)
	assert.Equal(t, cmd, session.Calls[0])
}
