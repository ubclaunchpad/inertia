package client

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigWrite(t *testing.T) {
	var writer bytes.Buffer
	config := &Config{
		CurrentRemoteName: "test",
		CurrentRemoteVPS:  getTestRemote(),
		Writer:            &writer,
	}
	inertiaJSON, err := json.Marshal(config)
	assert.Nil(t, err)

	n, err := config.Write()

	assert.Nil(t, err)
	assert.Equal(t, n, len(inertiaJSON))
}
