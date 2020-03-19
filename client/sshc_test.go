package client

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client/runner/mocks"
)

func newMockSSHClient(t *testing.T, m *mocks.FakeSSHSession) *Client {
	return &Client{
		ssh:   m,
		out:   &testWriter{t},
		debug: true,

		Remote: &cfg.Remote{
			Version: "test",
			IP:      "127.0.0.1",
			SSH: &cfg.SSH{
				IdentityFile: "../test/keys/id_rsa",
				User:         "root",
				SSHPort:      "69",
			},
			Daemon: &cfg.Daemon{
				Port: "4303",
			},
		},
	}
}

func TestSSHClient_InstallDocker(t *testing.T) {
	var session = &mocks.FakeSSHSession{}
	var client = newMockSSHClient(t, session)

	// Get original script for comparison
	script, err := ioutil.ReadFile("scripts/docker.sh")
	assert.NoError(t, err)

	// Get SSH runner
	sshc, err := client.GetSSHClient()
	assert.NoError(t, err)

	// Make sure the right command is run.
	assert.NoError(t, sshc.InstallDocker())
	call, interact := session.RunStreamArgsForCall(0)
	assert.False(t, interact)
	assert.Equal(t, string(script), call)
}

func TestSSHClient_DaemonUp(t *testing.T) {
	var session = &mocks.FakeSSHSession{}
	var client = newMockSSHClient(t, session)

	// Get original script for comparison
	script, err := ioutil.ReadFile("scripts/daemon-up.sh")
	assert.NoError(t, err)
	actualCommand := fmt.Sprintf(string(script), "test", "4303", "127.0.0.1", "")

	// Get SSH runner
	sshc, err := client.GetSSHClient()
	assert.NoError(t, err)

	// Make sure the right command is run.
	assert.NoError(t, sshc.DaemonUp())
	call, interact := session.RunStreamArgsForCall(0)
	assert.False(t, interact)
	assert.Equal(t, actualCommand, call)

	// Check with WEBHOOK_SECRET provided, make sure the right command is run.
	sshc.remote.Daemon.WebHookSecret = "sekret"
	assert.NoError(t, sshc.DaemonUp())
	call, interact = session.RunStreamArgsForCall(1)
	assert.False(t, interact)
	assert.Contains(t, call, "sekret")
}

func TestSSHClient_DaemonDown(t *testing.T) {
	var session = &mocks.FakeSSHSession{}
	var client = newMockSSHClient(t, session)

	// Get original script for comparison
	script, err := ioutil.ReadFile("scripts/daemon-down.sh")
	assert.NoError(t, err)

	// Get SSH runner
	sshc, err := client.GetSSHClient()
	assert.NoError(t, err)

	// Make sure the right command is run.
	assert.NoError(t, sshc.DaemonDown())
	assert.Equal(t, string(script), session.RunArgsForCall(0))
}

func TestSSHClient_KeyGen(t *testing.T) {
	var session = &mocks.FakeSSHSession{}
	var client = newMockSSHClient(t, session)

	// Get original script for comparison
	script, err := ioutil.ReadFile("scripts/token.sh")
	assert.NoError(t, err)
	tokenScript := fmt.Sprintf(string(script), "test")

	// Get SSH runner
	sshc, err := client.GetSSHClient()
	assert.NoError(t, err)

	// Make sure the right command is run.
	assert.NoError(t, sshc.AssignAPIToken())
	assert.Equal(t, tokenScript, session.RunArgsForCall(0))
}
