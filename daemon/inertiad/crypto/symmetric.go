package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

const (
	// SymmetricKeyLength is the length of the symmetric key in bytes
	SymmetricKeyLength = 32
)

// getRandomNonce returns a random nonce for AES GCM mode
func getRandomNonce(size int) []byte {
	nonce := make([]byte, size)
	rand.Read(nonce)
	return nonce
}

// Encrypt encrypts plaintext using given key in AES GCM mode
func Encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	nonce := getRandomNonce(aesgcm.NonceSize())

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// store nonce at beginning of ciphertext
	return append(nonce, ciphertext...), nil
}

// Decrypt decrypts ciphertext using given key in AES GCM mode
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()

	// nonce is stored at beginning of ciphertext
	nonce := ciphertext[:nonceSize]

	return aesgcm.Open(nil, nonce, ciphertext[nonceSize:], nil)
}
