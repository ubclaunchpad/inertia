package client

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockSSHRunner is a mocked out implementation of SSHSession
type mockSSHRunner struct {
	c     *Client
	Calls []string
}

func (runner *mockSSHRunner) Run(cmd string) (*bytes.Buffer, *bytes.Buffer, error) {
	runner.Calls = append(runner.Calls, cmd)
	return nil, nil, nil
}

func (runner *mockSSHRunner) RunStream(cmd string, interactive bool) error {
	runner.Calls = append(runner.Calls, cmd)
	return nil
}

func (runner *mockSSHRunner) RunSession() error {
	return nil
}

func TestRun(t *testing.T) {
	session := mockSSHRunner{c: getMockClient(nil)}
	cmd := "ls -lsa"

	_, _, err := session.Run(cmd)
	assert.Nil(t, err)
	assert.Equal(t, cmd, session.Calls[0])
}

func TestRunInteractive(t *testing.T) {
	session := mockSSHRunner{c: getMockClient(nil)}
	cmd := "ls -lsa"

	err := session.RunStream(cmd, true)
	assert.Nil(t, err)
	assert.Equal(t, cmd, session.Calls[0])
}
