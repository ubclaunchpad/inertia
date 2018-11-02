package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// KeyDerivationSaltLength is the length of the salt in bytes
	KeyDerivationSaltLength = 8
	// KeyDerivationKeyLength is the length of the key derived in bytes
	KeyDerivationKeyLength  = 32
	keyDerivationIterations = 10000
)

// GenerateSalt returns a random hex encoded salt for KD algorithm
func GenerateSalt() []byte {
	salt := make([]byte, KeyDerivationSaltLength)
	rand.Read(salt)
	return salt
}

// DeriveKey derives an AES encryption key based on salt + user's
// password using PBKDF2 with HMAC-SHA1
func DeriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key(
		[]byte(password),
		salt,
		keyDerivationIterations,
		KeyDerivationKeyLength,
		sha256.New,
	)
}
