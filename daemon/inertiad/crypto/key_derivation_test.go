package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateSalt(t *testing.T) {
	salt := GenerateSalt()
	assert.Equal(t, len(salt), KeyDerivationSaltLength)
}

func TestDeriveKey(t *testing.T) {
	password := "hunter7"
	salt := GenerateSalt()

	key := DeriveKey(password, salt)
	assert.Equal(t, len(key), KeyDerivationKeyLength)
}
