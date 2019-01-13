package cfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetProperty(t *testing.T) {
	testDaemonConfig := &DaemonConfig{
		Port:  "8080",
		Token: "abcdefg",
	}

	testRemote := &RemoteVPS{
		Name:   "testName",
		IP:     "1234",
		User:   "testUser",
		PEM:    "/some/pem/file",
		Daemon: testDaemonConfig,
	}
	a := SetProperty("name", "newTestName", testRemote)
	assert.True(t, a)
	assert.Equal(t, "newTestName", testRemote.Name)

	b := SetProperty("wrongtag", "otherTestName", testRemote)
	assert.False(t, b)
	assert.Equal(t, "newTestName", testRemote.Name)

	c := SetProperty("port", "8000", testDaemonConfig)
	assert.True(t, c)
	assert.Equal(t, "8000", testDaemonConfig.Port)
}
