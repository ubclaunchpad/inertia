package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestRemote() *RemoteVPS {
	return &RemoteVPS{
		IP:   "127.0.0.1",
		PEM:  "/Users/me/and/my/pem",
		User: "me",
		Port: "5555",
	}
}

func TestConfigWrite(t *testing.T) {
	config := &Config{
		CurrentRemoteName: "test",
		CurrentRemoteVPS:  getTestRemote(),
	}

	var f bytes.Buffer
	n, err := config.Write(&f)

	assert.Nil(t, err)
	assert.Equal(t, n, 98)
}
