package local

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/client"
)

func TestConfigCreateAndWriteAndRead(t *testing.T) {
	err := CreateConfigFile("test", "dockerfile", "")
	assert.Nil(t, err)
	config, configPath, err := GetProjectConfigFromDisk()
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
	err = config.Write(configPath)
	assert.Nil(t, err)

	readConfig, _, err := GetProjectConfigFromDisk()
	assert.Nil(t, err)
	assert.Equal(t, config.Remotes[0], readConfig.Remotes[0])
	assert.Equal(t, config.Remotes[1], readConfig.Remotes[1])

	err = os.Remove(configPath)
	assert.Nil(t, err)
}
