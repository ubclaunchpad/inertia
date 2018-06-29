package cfg

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig("test", "best-project", "docker-compose", "")
	assert.Equal(t, cfg.Version, "test")
}

func TestWriteFailed(t *testing.T) {
	cfg := NewConfig("test", "best-project", "docker-compose", "")
	err := cfg.Write("")
	assert.NotNil(t, err)
	assert.Contains(t, "nothing to write to", err.Error())
}

func TestWriteToPath(t *testing.T) {
	configPath := "/test-config.toml"
	cfg := NewConfig("test", "best-project", "docker-compose", "")

	cwd, err := os.Getwd()
	assert.Nil(t, err)
	absPath := filepath.Join(cwd, configPath)
	defer os.RemoveAll(absPath)

	err = cfg.Write(absPath)
	assert.Nil(t, err)

	writtenConfigContents, err := ioutil.ReadFile(absPath)
	assert.Nil(t, err)
	assert.Contains(t, string(writtenConfigContents), "best-project")
	assert.Contains(t, string(writtenConfigContents), "docker-compose")
}

func TestWriteToWritersAndFile(t *testing.T) {
	configPath := "/test-config.toml"
	cfg := NewConfig("test", "best-project", "docker-compose", "")

	cwd, err := os.Getwd()
	assert.Nil(t, err)
	absPath := filepath.Join(cwd, configPath)
	defer os.RemoveAll(absPath)

	buffer1 := bytes.NewBuffer(nil)
	buffer2 := bytes.NewBuffer(nil)

	err = cfg.Write(absPath, buffer1, buffer2)
	assert.Nil(t, err)

	writtenConfigContents, err := ioutil.ReadFile(absPath)
	assert.Nil(t, err)
	assert.Contains(t, string(writtenConfigContents), "best-project")
	assert.Contains(t, string(writtenConfigContents), "docker-compose")
	assert.Contains(t, buffer1.String(), "best-project")
	assert.Contains(t, buffer2.String(), "best-project")
}

func TestConfigGetRemote(t *testing.T) {
	config := &Config{Remotes: make(map[string]*RemoteVPS)}
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
	config := &Config{Remotes: make(map[string]*RemoteVPS)}
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

	added := config.AddRemote(testRemote)
	assert.False(t, added)

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
