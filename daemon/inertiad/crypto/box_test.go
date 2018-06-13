package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSealAndUndoSeal(t *testing.T) {
	encryptPublicKey, encryptPrivateKey, decryptPublicKey, decryptPrivateKey, err := GenerateKeys()
	assert.Nil(t, err)

	input := []byte("hello world")

	encrypted, err := Seal(input, encryptPrivateKey, decryptPublicKey)
	assert.Nil(t, err)
	assert.NotEqual(t, input, encrypted)

	// wrong keys
	_, err = UndoSeal(encrypted, encryptPrivateKey, decryptPublicKey)
	assert.NotNil(t, err)

	// Successfully undo seal
	decrypted, err := UndoSeal(encrypted, encryptPublicKey, decryptPrivateKey)
	assert.Nil(t, err)
	assert.Equal(t, input, decrypted)
}
