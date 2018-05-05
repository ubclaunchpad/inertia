package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/client"
)

func TestConfigCreateAndWriteAndRead(t *testing.T) {
	err := createConfigFile("", "")
	assert.Nil(t, err)
	config, path, err := getProjectConfigFromDisk()
	assert.Nil(t, err)
	config.AddRemote(&client.RemoteVPS{
		Name:    "test",
		IP:      "1234",
		User:    "bobheadxi",
		PEM:     "/some/pem/file",
		SSHPort: "22",
		Daemon: &client.DaemonConfig{
			Port: "8080",
		},
	})
	config.AddRemote(&client.RemoteVPS{
		Name:    "test2",
		IP:      "12343",
		User:    "bobheadxi234",
		PEM:     "/some/pem/file234",
		SSHPort: "222",
		Daemon: &client.DaemonConfig{
			Port: "80801",
		},
	})
	err = config.Write(path)
	assert.Nil(t, err)

	readConfig, _, err := getProjectConfigFromDisk()
	assert.Nil(t, err)
	assert.Equal(t, config.GetRemotes()[0], readConfig.GetRemotes()[0])
	assert.Equal(t, config.GetRemotes()[1], readConfig.GetRemotes()[1])

	path, err = getConfigFilePath()
	assert.Nil(t, err)
	err = os.Remove(path)
	assert.Nil(t, err)
}
