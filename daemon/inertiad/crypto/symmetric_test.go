package crypto

import (
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {

	keyGood := make([]byte, SymmetricKeyLength)
	keyBad := make([]byte, SymmetricKeyLength)
	rand.Read(keyGood)
	rand.Read(keyBad)

	plaintext := []byte("I'm a little teapot, short and STDOUT")

	ciphertext, err := Encrypt(keyGood, plaintext)
	assert.NoError(t, err)

	decrypted, err := Decrypt(keyGood, ciphertext)
	assert.NoError(t, err)

	// Decrypted matches plaintext
	for i := range decrypted {
		assert.Equal(t, plaintext[i], decrypted[i])
	}

	// Bad key, should err here because MAC will be invalid
	_, err = Decrypt(keyBad, ciphertext)
	assert.NotNil(t, err)
}
