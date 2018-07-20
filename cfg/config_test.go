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
	cfg := NewConfig("test", "best-project", "docker-compose", "", "")
	assert.Equal(t, cfg.Version, "test")
}

func TestNewConfigFromFiles(t *testing.T) {
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	type args struct {
		project string
		remote  string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{"both files do not exist", args{"asdf", "asdf"}, nil, true},
		{"project is malformed", args{"inertia.go", ""}, nil, true},
		{"remotes is malformed", args{"", "inertia.go"}, nil, true},
		{
			"both files valid",
			args{"example.inertia.toml", "example.inertia.remotes"},
			&Config{Version: "latest", BuildType: "dockerfile"},
			false,
		},
		{
			"only project is valid",
			args{"example.inertia.toml", "asdf"},
			&Config{Version: "latest", BuildType: "dockerfile"},
			false,
		},
		{
			"only remotes is valid",
			args{"asdf", "example.inertia.remotes"},
			&Config{Version: "latest", BuildType: ""},
			false,
		},
	}
	for _, tt := range tests {
		// Set paths relative to root of project
		tt.args.project = filepath.Join(filepath.Dir(cwd), tt.args.project)
		tt.args.remote = filepath.Join(filepath.Dir(cwd), tt.args.remote)
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfigFromFiles(tt.args.project, tt.args.remote)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfigFromFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, got.Version, tt.want.Version)
				assert.Equal(t, got.BuildType, tt.want.BuildType)
			}
		})
	}
}

func TestWriteFailed(t *testing.T) {
	cfg := NewConfig("test", "best-project", "docker-compose", "", "")
	err := cfg.WriteProjectConfig("")
	assert.NotNil(t, err)
	assert.Contains(t, "nothing to write to", err.Error())
}

func TestWriteProjectConfigToPath(t *testing.T) {
	configPath := "/test-config.toml"
	cfg := NewConfig("test", "best-project", "docker-compose", "", "")

	cwd, err := os.Getwd()
	assert.Nil(t, err)
	absPath := filepath.Join(cwd, configPath)
	defer os.RemoveAll(absPath)

	err = cfg.WriteProjectConfig(absPath)
	assert.Nil(t, err)

	writtenConfigContents, err := ioutil.ReadFile(absPath)
	assert.Nil(t, err)
	assert.Contains(t, string(writtenConfigContents), "best-project")
	assert.Contains(t, string(writtenConfigContents), "docker-compose")
}

func TestWritePremoteConfigToPath(t *testing.T) {
	configPath := "/test-config.toml"
	cfg := NewConfig("test", "best-project", "docker-compose", "", "")
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
	cfg.AddRemote(testRemote)

	cwd, err := os.Getwd()
	assert.Nil(t, err)
	absPath := filepath.Join(cwd, configPath)
	defer os.RemoveAll(absPath)

	err = cfg.WriteRemoteConfig(absPath)
	assert.Nil(t, err)

	writtenConfigContents, err := ioutil.ReadFile(absPath)
	assert.Nil(t, err)
	assert.Contains(t, string(writtenConfigContents), "/some/pem/file")
	assert.Contains(t, string(writtenConfigContents), "bobheadxi")
}

func TestWriteToWritersAndFile(t *testing.T) {
	configPath := "/test-config.toml"
	cfg := NewConfig("test", "best-project", "docker-compose", "", "")

	cwd, err := os.Getwd()
	assert.Nil(t, err)
	absPath := filepath.Join(cwd, configPath)
	defer os.RemoveAll(absPath)

	buffer1 := bytes.NewBuffer(nil)
	buffer2 := bytes.NewBuffer(nil)

	err = cfg.WriteProjectConfig(absPath, buffer1, buffer2)
	assert.Nil(t, err)

	writtenConfigContents, err := ioutil.ReadFile(absPath)
	assert.Nil(t, err)
	assert.Contains(t, string(writtenConfigContents), "best-project")
	assert.Contains(t, string(writtenConfigContents), "docker-compose")
	assert.Contains(t, buffer1.String(), "best-project")
	assert.Contains(t, buffer2.String(), "best-project")
}

func TestConfigGetRemote(t *testing.T) {
	config := &Config{remotes: make(map[string]*RemoteVPS)}
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
	config := &Config{remotes: make(map[string]*RemoteVPS)}
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
