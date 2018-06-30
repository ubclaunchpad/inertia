package crypto

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	TokenInvalidErrorMsg = "token invalid"
	TokenExpiredErrorMsg = "token expired"
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
	// TODO: This is a workaround for client tokens, which currently do not
	// have claims. Need to move those onto this system.
	if t.SessionID == "" && t.User == "" {
		return nil
	}

	if !t.Expiry.After(time.Now()) {
		return errors.New(TokenExpiredErrorMsg)
	}
	return nil
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
		return nil, errors.New(TokenInvalidErrorMsg)
	}

	// Verify the claims and token.
	if claim, ok := token.Claims.(*TokenClaims); ok {
		return claim, nil
	}
	return nil, errors.New(TokenInvalidErrorMsg)
}

// GenerateToken creates a JSON Web Token (JWT) for a client to use when
// sending HTTP requests to the daemon server.
func GenerateToken(key []byte) (string, error) {
	// No claims for now.
	return jwt.New(jwt.SigningMethodHS256).SignedString(key)
}
