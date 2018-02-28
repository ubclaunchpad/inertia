package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigCreateAndWriteAndRead(t *testing.T) {
	err := createConfigDirectory()
	assert.Nil(t, err)
	config, err := GetProjectConfigFromDisk()
	assert.Nil(t, err)
	config.Remotes["test"] = &RemoteVPS{
		IP:   "1234",
		User: "bobheadxi",
		PEM:  "/some/pem/file",
		Daemon: &DaemonConfig{
			Port:    "8080",
			SSHPort: "22",
		},
	}
	err = config.Write()
	assert.Nil(t, err)

	readConfig, err := GetProjectConfigFromDisk()
	assert.Nil(t, err)
	assert.Equal(t, config.Remotes["test"], readConfig.Remotes["test"])

	cwd, _ := os.Getwd()
	os.Remove(filepath.Join(cwd, configFileName))
}
