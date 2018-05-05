package client

import (
	"bytes"
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

func (runner *mockSSHRunner) RunStream(cmd string, interactive bool) error {
	runner.Calls = append(runner.Calls, cmd)
	return nil
}

func (runner *mockSSHRunner) RunSession() error {
	return nil
}
