package provision

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEC2Provisioner(t *testing.T) {
	prov, _ := NewEC2Provisioner("id", "key")
	assert.NotNil(t, prov.client.Config.Credentials)
}

func TestNewEC2ProvisionerFromEnv(t *testing.T) {
	prov, _ := NewEC2Provisioner("id", "key")
	assert.NotNil(t, prov.client.Config.Credentials)
}
