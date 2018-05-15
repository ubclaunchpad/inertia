package client

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigCreateAndWriteAndRead(t *testing.T) {
	err := createConfigFile("", "")
	assert.Nil(t, err)
	config, err := GetProjectConfigFromDisk()
	assert.Nil(t, err)
	config.AddRemote(&RemoteVPS{
		Name: "test",
		IP:   "1234",
		User: "bobheadxi",
		PEM:  "/some/pem/file",
		Daemon: &DaemonConfig{
			Port:    "8080",
			SSHPort: "22",
		},
	})
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
	err = config.Write()
	assert.Nil(t, err)

	readConfig, err := GetProjectConfigFromDisk()
	assert.Nil(t, err)
	assert.Equal(t, config.Remotes[0], readConfig.Remotes[0])
	assert.Equal(t, config.Remotes[1], readConfig.Remotes[1])

	path, err := GetConfigFilePath()
	assert.Nil(t, err)
	println(path)
	err = os.Remove(path)
	assert.Nil(t, err)
}

func TestConfigGetRemote(t *testing.T) {
	config := &Config{Remotes: make([]*RemoteVPS, 0)}
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
	config := &Config{Remotes: make([]*RemoteVPS, 0)}
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

func TestSetProperty(t *testing.T) {

	testDaemonConfig := &DaemonConfig{
		Port:    "8080",
		SSHPort: "22",
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
