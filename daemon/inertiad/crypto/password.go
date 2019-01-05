package crypto

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	errSameUsernamePassword = errors.New("Username and password must be different")
	errInvalidUsername      = errors.New("Username must be at least 3 characters and only letters, numbers, underscores, and dashes are allowed")
	errInvalidPassword      = errors.New("Password must be at least 5 characters and only letters, numbers, underscores, and dashes are allowed")
)

// IsCredentialFormatError returns true if the given error is one related to
// username/password format
func IsCredentialFormatError(err error) bool {
	return strings.Contains(err.Error(), errSameUsernamePassword.Error()) ||
		strings.Contains(err.Error(), errInvalidUsername.Error()) ||
		strings.Contains(err.Error(), errInvalidPassword.Error())
}

// HashPassword generates a bcrypt-encrypted hash from given password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("bcrypt password hashing unsuccessful: " + err.Error())
	}
	return string(hash), nil
}

// CorrectPassword checks if given password maps correctly to the given hash
func CorrectPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ValidateCredentialValues takes a username and password and verifies
// if they are of sufficient length and if they only contain legal characters
func ValidateCredentialValues(username, password string) error {
	println(username, password)
	if username == password {
		return errSameUsernamePassword
	}
	if len(password) < 5 || len(password) >= 128 || !IsLegalString(password) {
		return errInvalidPassword
	}
	if len(username) < 3 || len(username) >= 128 || !IsLegalString(username) {
		return errInvalidUsername
	}
	return nil
}

// IsLegalString returns true if `str` only contains characters [A-Z], [a-z], or '_' or '-'
func IsLegalString(str string) bool {
	for _, c := range str {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < 48 || c > 57) && c != '_' && c != '-' {
			return false
		}
	}
	return true
}
