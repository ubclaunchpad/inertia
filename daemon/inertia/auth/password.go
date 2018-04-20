package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	errSameUsernamePassword = errors.New("Username and password must be different")
	errInvalidUsername      = errors.New("Only letters, numbers, underscores, and dashes are allowed in usernames, and username must be at least 3 characters")
	errInvalidPassword      = errors.New("Only letters, numbers, underscores, and dashes are allowed in passwords, and password must be at least 5 characters")
)

// hashPassword generates a bcrypt-encrypted hash from given password
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("bcrypt password hashing unsuccessful: " + err.Error())
	}
	return string(hash), nil
}

// correctPassword checks if given password maps correctly to the given hash
func correctPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// validateCredentialValues takes a username and password and verifies
// if they are of sufficient length and if they only contain legal characters
func validateCredentialValues(username, password string) error {
	if username == password {
		return errSameUsernamePassword
	}
	if len(password) < 5 || len(password) >= 128 || !isLegalString(password) {
		return errInvalidPassword
	}
	if len(username) < 3 || len(username) >= 128 || !isLegalString(username) {
		return errInvalidUsername
	}
	return nil
}

// isLegalString returns true if `str` only contains characters [A-Z], [a-z], or '_' or '-'
func isLegalString(str string) bool {
	for _, c := range str {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && c != '_' && c != '-' {
			return false
		}
	}
	return true
}
