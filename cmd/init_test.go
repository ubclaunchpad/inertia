package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
