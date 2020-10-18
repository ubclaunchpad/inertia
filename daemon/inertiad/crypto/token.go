package crypto

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	// ErrInvalidToken says that the token is invalid
	ErrInvalidToken = errors.New("token invalid")

	// ErrTokenExpired says that the token is expired
	ErrTokenExpired = errors.New("token expired")
)

// TokenClaims represents a JWT token's claims
type TokenClaims struct {
	SessionID string    `json:"session_id"`
	User      string    `json:"user"`
	Admin     bool      `json:"admin"`
	Expiry    time.Time `json:"expiry"`
}

// Valid checks if token is authentic
func (t *TokenClaims) Valid() error {
	if t.IsMaster() {
		return nil
	}

	if !t.Expiry.After(time.Now()) {
		return ErrTokenExpired
	}
	return nil
}

// IsMaster returns true if this is a master key
func (t *TokenClaims) IsMaster() bool {
	return (t.User == "master" && t.Expiry == time.Time{})
}

// GenerateToken creates a JWT token from this claim, signed with given key
func (t *TokenClaims) GenerateToken(key []byte) (string, error) {
	return jwt.
		NewWithClaims(jwt.SigningMethodHS256, t).
		SignedString(key)
}

// ValidateToken ensures token is valid and returns its metadata
func ValidateToken(tokenString string, lookup jwt.Keyfunc) (*TokenClaims, error) {
	// Parse takes the token string and a function for looking up the key.
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, lookup)
	if err != nil {
		return nil, err
	}

	// Verify signing algorithm and token
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify the claims and token.
	if claim, ok := token.Claims.(*TokenClaims); ok {
		return claim, nil
	}
	return nil, ErrInvalidToken
}

// GenerateMasterToken creates a "master" JSON Web Token (JWT) for a client to use
// when sending HTTP requests to the daemon server.
func GenerateMasterToken(key []byte) (string, error) {
	return jwt.
		NewWithClaims(jwt.SigningMethodHS256, &TokenClaims{
			User:  "master",
			Admin: true,
			// For the time being, never allow this token to expire, so don't
			// set an expiry.
		}).
		SignedString(key)
}
