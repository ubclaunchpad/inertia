package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	errSameUsernamePassword = errors.New("Username and password must be different")
	errInvalidUsername      = errors.New("Only letters, numbers and underscores are allowed in usernames")
	errInvalidPassword      = errors.New("Only letters, numbers and underscores are allowed in passwords, and password must be at least 5 characters")
)

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("bcrypt password hashing unsuccessful: " + err.Error())
	}
	return string(hash), nil
}

func correctPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// validateCredentialValues takes a username and password and verifies
// if they are of sufficient length and if they only contain legal characters
func validateCredentialValues(username, password string) error {
	if username == password {
		return errSameUsernamePassword
	}
	if len(password) < 5 || len(password) >= 128 {
		return errInvalidPassword
	}
	if len(username) < 3 || len(username) >= 128 {
		return errInvalidUsername
	}
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

NEXT_USERNAME_CHAR:
	for _, char := range username {
		for _, validChar := range validChars {
			// if valid, skip to next character
			if char == validChar {
				continue NEXT_USERNAME_CHAR
			}
		}
		return errInvalidUsername
	}

NEXT_PASSWORD_CHAR:
	for _, char := range password {
		for _, validChar := range validChars {
			// if valid, skip to next character
			if char == validChar {
				continue NEXT_PASSWORD_CHAR
			}
		}
		return errInvalidPassword
	}

	return nil
}
