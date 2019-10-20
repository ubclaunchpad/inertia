package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testInertiaKeyPath = "../../../test/keys/id_rsa"

func TestGitAuthFailedErr(t *testing.T) {
	err := AuthFailedErr(testInertiaKeyPath)
	assert.NotNil(t, err)
	// Check for a substring of public key
	assert.Contains(t, err.Error(), "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDD")
}

func TestGitAuthFailedErrFailed(t *testing.T) {
	err := AuthFailedErr("wow")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Error reading key")
}
