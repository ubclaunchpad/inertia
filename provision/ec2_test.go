package provision

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEC2Provisioner(t *testing.T) {
	prov, _ := NewEC2Provisioner("bob", "id", "key")
	assert.NotNil(t, prov.client.Config.Credentials)
	assert.Equal(t, "bob", prov.GetUser())
}

func TestNewEC2ProvisionerFromEnv(t *testing.T) {
	prov, _ := NewEC2Provisioner("bob", "id", "key")
	assert.NotNil(t, prov.client.Config.Credentials)
	assert.Equal(t, "bob", prov.GetUser())
}

func TestNewEC2ProvisionerFromProfile(t *testing.T) {
	prov, _ := NewEC2ProvisionerFromProfile("bob", "", "../test/aws/credentials")
	assert.NotNil(t, prov.client.Config.Credentials)
	assert.Equal(t, "bob", prov.GetUser())
}
