package cfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemote_GetProfile(t *testing.T) {
	var remote = &Remote{IP: "127.0.0.1"}
	assert.Equal(t, "default", remote.GetProfile("boring"))

	remote.Profiles = map[string]string{"chicken": "drumsticks"}
	assert.Equal(t, "drumsticks", remote.GetProfile("chicken"))
}

func TestRemote_ApplyProfile(t *testing.T) {
	var remote = &Remote{IP: "127.0.0.1"}
	remote.ApplyProfile("hot", "wings")
	remote.ApplyProfile("chicken", "drumsticks")
	assert.Equal(t, "wings", remote.GetProfile("hot"))
	assert.Equal(t, "drumsticks", remote.GetProfile("chicken"))
}

func TestRemote_GetHost(t *testing.T) {
	var remote = &Remote{IP: "127.0.0.1"}
	_, err := remote.SSHHost()
	assert.Error(t, err)

	remote.SSH = &SSH{User: "bobheadxi"}
	host, err := remote.SSHHost()
	assert.NoError(t, err)
	assert.Equal(t, "bobheadxi@127.0.0.1", host)
}

func TestRemote_GetIPAndPort(t *testing.T) {
	var remote = &Remote{IP: "127.0.0.1"}
	_, err := remote.DaemonAddr()
	assert.Error(t, err)

	remote.Daemon = &Daemon{Port: "4303"}
	addr, err := remote.DaemonAddr()
	assert.Equal(t, "https://127.0.0.1:4303", addr)
}
