package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/cfg"
)

func TestSetProperty(t *testing.T) {
	testDaemonConfig := &cfg.DaemonConfig{
		Port:  "8080",
		Token: "abcdefg",
	}

	testRemote := &cfg.RemoteVPS{
		Name:   "testName",
		IP:     "1234",
		User:   "testUser",
		PEM:    "/some/pem/file",
		Daemon: testDaemonConfig,
	}
	a := setProperty("name", "newTestName", testRemote)
	assert.True(t, a)
	assert.Equal(t, "newTestName", testRemote.Name)

	b := setProperty("wrongtag", "otherTestName", testRemote)
	assert.False(t, b)
	assert.Equal(t, "newTestName", testRemote.Name)

	c := setProperty("port", "8000", testDaemonConfig)
	assert.True(t, c)
	assert.Equal(t, "8000", testDaemonConfig.Port)
}
