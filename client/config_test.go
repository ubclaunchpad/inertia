package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigGetRemote(t *testing.T) {
	config := &Config{remotes: make([]*RemoteVPS, 0)}
	testRemote := &RemoteVPS{
		Name: "test",
		IP:   "12343",
		User: "bobheadxi",
		PEM:  "/some/pem/file",
		Daemon: &DaemonConfig{
			Port:    "8080",
			SSHPort: "22",
		},
	}
	config.AddRemote(testRemote)
	remote, found := config.GetRemote("test")
	assert.True(t, found)
	assert.Equal(t, testRemote, remote)

	_, found = config.GetRemote("what")
	assert.False(t, found)
}

func TestConfigRemoteRemote(t *testing.T) {
	config := &Config{remotes: make([]*RemoteVPS, 0)}
	testRemote := &RemoteVPS{
		Name: "test",
		IP:   "12343",
		User: "bobheadxi",
		PEM:  "/some/pem/file",
		Daemon: &DaemonConfig{
			Port:    "8080",
			SSHPort: "22",
		},
	}
	config.AddRemote(testRemote)
	config.AddRemote(&RemoteVPS{
		Name: "test2",
		IP:   "12343",
		User: "bobheadxi234",
		PEM:  "/some/pem/file234",
		Daemon: &DaemonConfig{
			Port:    "80801",
			SSHPort: "222",
		},
	})
	removed := config.RemoveRemote("test2")
	assert.True(t, removed)
	removed = config.RemoveRemote("what")
	assert.False(t, removed)

	remote, found := config.GetRemote("test")
	assert.True(t, found)
	assert.Equal(t, testRemote, remote)
}
