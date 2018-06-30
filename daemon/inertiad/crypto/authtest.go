package crypto

import (
	"os"
	"path"

	jwt "github.com/dgrijalva/jwt-go"
)

// This file contains test assets

var (
	// TestPrivateKey is an example key for testing purposes
	TestPrivateKey = []byte("very_sekrit_key")

	// TestToken is an example token for testing purposes
	TestToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.AqFWnFeY9B8jj7-l3z0a9iaZdwIca7xhUF3fuaJjU90"

	// TestInertiaKeyPath the path to Inertia's test RSA key
	TestInertiaKeyPath = path.Join(os.Getenv("GOPATH"), "/src/github.com/ubclaunchpad/inertia/test/keys/id_rsa")
)

// GetFakeAPIKey is a helper function that implements jwt.keyFunc and returns
// the test private key
func GetFakeAPIKey(tok *jwt.Token) (interface{}, error) {
	return TestPrivateKey, nil
}
