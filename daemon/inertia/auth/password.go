package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	errSameUsernamePassword = errors.New("Username and password must be different")
	errInvalidUsername      = errors.New("Only letters, numbers and underscore are allowed in usernames")
	errInvalidPassword      = errors.New("Only letters, numbers and underscore are allowed in passwords, and password must be at least 5 characters")
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

func validateCredentialValues(username, password string) error {
	if username == password {
		return errSameUsernamePassword
	}
	if len(password) < 5 {
		return errInvalidPassword
	}
	validChars := "abcdefghijklmnopqrstuvwxyzæøåABCDEFGHIJKLMNOPQRSTUVWXYZÆØÅ_0123456789"

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
