package crypto

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"strings"
)

const (
	// Prefixes used by GitHub before the HMAC hexdigest.
	sha1Prefix = "sha1"
)

// ValidateSignature validates the HMAC signature for the given payload.
// Based off of https://github.com/google/go-github
func ValidateSignature(signature string, payload, secretKey []byte) error {
	messageMAC, hashFunc, err := messageMAC(signature)
	if err != nil {
		return err
	}
	if !checkMAC(payload, messageMAC, secretKey, hashFunc) {
		return errors.New("payload signature check failed")
	}
	return nil
}

// checkMAC reports whether messageMAC is a valid HMAC tag for message.
func checkMAC(message, messageMAC, key []byte, hashFunc func() hash.Hash) bool {
	mac := hmac.New(hashFunc, key)
	mac.Write(message)
	return hmac.Equal(messageMAC, mac.Sum(nil))
}

// messageMAC returns the hex-decoded HMAC tag from the signature and its
// corresponding hash function.
func messageMAC(signature string) ([]byte, func() hash.Hash, error) {
	if signature == "" {
		return nil, nil, errors.New("missing signature")
	}
	sigParts := strings.SplitN(signature, "=", 2)
	if len(sigParts) != 2 {
		return nil, nil, fmt.Errorf("error parsing signature %q", signature)
	}

	var (
		signaturePrefix  = sigParts[0]
		payloadSignature = sigParts[1]
		hashFunc         func() hash.Hash
	)

	switch signaturePrefix {
	case sha1Prefix:
		hashFunc = sha1.New
	default:
		return nil, nil, fmt.Errorf("unknown hash type prefix: %q", signaturePrefix)
	}

	buf, err := hex.DecodeString(payloadSignature)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding signature %q: %v", signature, err)
	}
	return buf, hashFunc, nil
}
