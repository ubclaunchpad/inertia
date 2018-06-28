package provision

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEC2Provisioner(t *testing.T) {
	_, err := NewEC2Provisioner("id", "key")
	assert.NotNil(t, err)
}

func TestNewEC2ProvisionerFromEnv(t *testing.T) {
	_, err := NewEC2Provisioner("id", "key")
	assert.NotNil(t, err)
}
