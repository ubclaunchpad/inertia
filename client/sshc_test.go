package client

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/client/runner/mocks"
)

func TestInstallDocker(t *testing.T) {
	var session = &mocks.FakeSSHSession{}
	var client = newMockSSHClient(session)

	// Get original script for comparison
	script, err := ioutil.ReadFile("scripts/docker.sh")
	assert.NoError(t, err)

	// Get SSH runner
	sshc, err := client.GetSSHClient()
	assert.NoError(t, err)

	// Make sure the right command is run.
	assert.NoError(t, sshc.InstallDocker())
	assert.Equal(t, string(script), session.RunArgsForCall(0))
}

func TestDaemonUp(t *testing.T) {
	var session = &mocks.FakeSSHSession{}
	var client = newMockSSHClient(session)

	// Get original script for comparison
	script, err := ioutil.ReadFile("scripts/daemon-up.sh")
	assert.NoError(t, err)
	actualCommand := fmt.Sprintf(string(script), "test", "4303", "127.0.0.1")

	// Get SSH runner
	sshc, err := client.GetSSHClient()
	assert.NoError(t, err)

	// Make sure the right command is run.
	assert.NoError(t, sshc.DaemonUp())
	assert.Equal(t, actualCommand, session.RunArgsForCall(0))
}

func TestDaemonDown(t *testing.T) {
	var session = &mocks.FakeSSHSession{}
	var client = newMockSSHClient(session)

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

func TestKeyGen(t *testing.T) {
	var session = &mocks.FakeSSHSession{}
	var client = newMockSSHClient(session)

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
