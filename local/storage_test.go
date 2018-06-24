package local

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/cfg"
)

func TestInitializeInertiaProjetFail(t *testing.T) {
	err := InitializeInertiaProject("inertia.toml", "", "", "")
	assert.NotNil(t, err)
}

func TestGetConfigFail(t *testing.T) {
	_, _, err := GetProjectConfigFromDisk("inertia.toml")
	assert.NotNil(t, err)
}

func TestConfigCreateAndWriteAndRead(t *testing.T) {
	err := createConfigFile("inertia.toml", "test", "dockerfile", "")
	assert.Nil(t, err)

	// Already exists
	err = createConfigFile("inertia.toml", "test", "dockerfile", "")
	assert.NotNil(t, err)

	// Get config and add remotes
	config, configPath, err := GetProjectConfigFromDisk("inertia.toml")
	assert.Nil(t, err)
	config.AddRemote(&cfg.RemoteVPS{
		Name:    "test",
		IP:      "1234",
		User:    "bobheadxi",
		PEM:     "/some/pem/file",
		SSHPort: "22",
		Daemon: &cfg.DaemonConfig{
			Port: "8080",
		},
	})
	config.AddRemote(&cfg.RemoteVPS{
		Name:    "test2",
		IP:      "12343",
		User:    "bobheadxi234",
		PEM:     "/some/pem/file234",
		SSHPort: "222",
		Daemon: &cfg.DaemonConfig{
			Port: "80801",
		},
	})

	// Test config creation
	err = config.Write(configPath)
	assert.Nil(t, err)

	// Test config read
	readConfig, _, err := GetProjectConfigFromDisk("inertia.toml")
	assert.Nil(t, err)
	assert.Equal(t, config.Remotes[0], readConfig.Remotes[0])
	assert.Equal(t, config.Remotes[1], readConfig.Remotes[1])

	// Test client read
	client, err := GetClient("test2", "inertia.toml")
	assert.Nil(t, err)
	assert.Equal(t, "test2", client.Name)
	assert.Equal(t, "12343:80801", client.GetIPAndPort())
	_, err = GetClient("asdf", "inertia.toml")
	assert.NotNil(t, err)

	// Test config remove
	err = os.Remove(configPath)
	assert.Nil(t, err)
}
