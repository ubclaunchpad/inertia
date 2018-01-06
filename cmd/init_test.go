package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigWrite(t *testing.T) {
	config := &Config{
		CurrentRemoteName: "test",
		CurrentRemoteVPS:  getTestRemote(),
	}
	inertiaJSON, err := json.Marshal(config)
	assert.Nil(t, err)

	var f bytes.Buffer
	n, err := config.Write(&f)

	assert.Nil(t, err)
	assert.Equal(t, n, len(inertiaJSON))
}
