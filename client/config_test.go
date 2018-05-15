package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigGetRemote(t *testing.T) {
	config := &Config{Remotes: make([]*RemoteVPS, 0)}
	testRemote := &RemoteVPS{
		Name:    "test",
		IP:      "12343",
		User:    "bobheadxi",
		PEM:     "/some/pem/file",
		SSHPort: "22",
		Daemon: &DaemonConfig{
			Port: "8080",
		},
	}
	config.AddRemote(testRemote)
	remote, found := config.GetRemote("test")
	assert.True(t, found)
	assert.Equal(t, testRemote, remote)

	_, found = config.GetRemote("what")
	assert.False(t, found)
}

func TestConfigRemoveRemote(t *testing.T) {
	config := &Config{Remotes: make([]*RemoteVPS, 0)}
	testRemote := &RemoteVPS{
		Name:    "test",
		IP:      "12343",
		User:    "bobheadxi",
		PEM:     "/some/pem/file",
		SSHPort: "22",
		Daemon: &DaemonConfig{
			Port: "8080",
		},
	}
	config.AddRemote(testRemote)
	config.AddRemote(&RemoteVPS{
		Name:    "test2",
		IP:      "12343",
		User:    "bobheadxi234",
		PEM:     "/some/pem/file234",
		SSHPort: "222",
		Daemon: &DaemonConfig{
			Port: "80801",
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

func TestSetProperty(t *testing.T) {
	testDaemonConfig := &DaemonConfig{
		Port:  "8080",
		Token: "abcdefg",
	}

	testRemote := &RemoteVPS{
		Name:   "testName",
		IP:     "1234",
		User:   "testUser",
		PEM:    "/some/pem/file",
		Daemon: testDaemonConfig,
	}
	a := SetProperty("name", "newTestName", testRemote)
	assert.True(t, a)
	assert.Equal(t, "newTestName", testRemote.Name)

	b := SetProperty("wrongtag", "otherTestName", testRemote)
	assert.False(t, b)
	assert.Equal(t, "newTestName", testRemote.Name)

	c := SetProperty("port", "8000", testDaemonConfig)
	assert.True(t, c)
	assert.Equal(t, "8000", testDaemonConfig.Port)
}
