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

func TestListImageOptionsNoAuth(t *testing.T) {
	prov, err := NewEC2Provisioner("id", "key")
	assert.NotNil(t, prov.client.Config.Credentials)

	_, err = prov.ListImageOptions("us-east-1")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "AuthFailure")
}
