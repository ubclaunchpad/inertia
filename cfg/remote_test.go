package cfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteVPS_GetHost(t *testing.T) {
	var remote = &Remote{IP: "127.0.0.1"}
	_, err := remote.SSHHost()
	assert.Error(t, err)

	remote.SSH = &SSH{User: "bobheadxi"}
	host, err := remote.SSHHost()
	assert.NoError(t, err)
	assert.Equal(t, "bobheadxi@127.0.0.1", host)
}

func TestRemoteVPS_GetIPAndPort(t *testing.T) {
	var remote = &Remote{IP: "127.0.0.1"}
	_, err := remote.DaemonAddr()
	assert.Error(t, err)

	remote.Daemon = &Daemon{Port: "4303"}
	addr, err := remote.DaemonAddr()
	assert.Equal(t, "https://127.0.0.1:4303", addr)
}
