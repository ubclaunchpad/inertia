package crypto

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// GenerateKeys creates 2 sets of keys - one for decryption, one for encryption
func GenerateKeys() (encryptPublicKey *[32]byte, encryptPrivateKey *[32]byte,
	decryptPublicKey *[32]byte, decryptPrivateKey *[32]byte, err error) {
	encryptPublicKey, encryptPrivateKey, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	decryptPublicKey, decryptPrivateKey, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return encryptPublicKey, encryptPrivateKey, decryptPublicKey, decryptPrivateKey, err
}

// Seal encrypts given value
func Seal(valueBytes []byte, encryptPrivateKey, decryptPublicKey *[32]byte) ([]byte, error) {
	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	// This encrypts the variable, storing the nonce in the first 24 bytes.
	variable := []byte(valueBytes)
	return box.Seal(
		nonce[:], variable, &nonce,
		decryptPublicKey, encryptPrivateKey,
	), nil
}

// UndoSeal decrypts sealed value
func UndoSeal(value []byte, encryptPublicKey, decryptPrivateKey *[32]byte) ([]byte, error) {
	// Decrypt the message using decrypt private key and the
	// encrypt public key. When you decrypt, you must use the same
	// nonce you used to encrypt the message - this nonce is stored
	// in the first 24 bytes.
	var decryptNonce [24]byte
	copy(decryptNonce[:], value[:24])
	decrypted, ok := box.Open(
		nil, value[24:], &decryptNonce,
		encryptPublicKey, decryptPrivateKey,
	)
	if !ok {
		return nil, errors.New("decryption error")
	}
	return decrypted, nil
}
