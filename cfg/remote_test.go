package cfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteVPS_GetHost(t *testing.T) {
	remote := &RemoteVPS{
		User: "bobheadxi",
		IP:   "127.0.0.1",
	}
	assert.Equal(t, "bobheadxi@127.0.0.1", remote.GetHost())
}

func TestRemoteVPS_GetIPAndPort(t *testing.T) {
	remote := &RemoteVPS{
		IP:     "127.0.0.1",
		Daemon: &DaemonConfig{Port: "4303"},
	}
	assert.Equal(t, "127.0.0.1:4303", remote.GetIPAndPort())
}
