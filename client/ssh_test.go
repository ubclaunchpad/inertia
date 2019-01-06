package client

import (
	"bytes"
	"io"

	"github.com/ubclaunchpad/inertia/cfg"
)

// mockSSHRunner is a mocked out implementation of SSHSession
type mockSSHRunner struct {
	r     *cfg.RemoteVPS
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

func (runner *mockSSHRunner) RunSession(commands ...string) error {
	return nil
}

func (runner *mockSSHRunner) CopyFile(f io.Reader, remotePath string, permissions string) error {
	return nil
}
