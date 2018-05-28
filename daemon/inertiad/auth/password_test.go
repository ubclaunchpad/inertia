package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	unhashed := "amazing"
	hashed, err := hashPassword(unhashed)
	assert.Nil(t, err)
	assert.NotEqual(t, unhashed, hashed)
}

func TestCorrectPassword(t *testing.T) {
	unhashed := "amazing"
	hashed, err := hashPassword(unhashed)
	assert.Nil(t, err)
	assert.NotEqual(t, unhashed, hashed)

	correct := correctPassword(hashed, unhashed)
	assert.True(t, correct)

	correct = correctPassword(hashed, "ummmmm")
	assert.False(t, correct)
}

func TestValidateCredentialValues(t *testing.T) {
	err := validateCredentialValues("finasdfsdfe", "okaasdfasdy")
	assert.Nil(t, err)

	err = validateCredentialValues("0123456789a", "0123456789")
	assert.Nil(t, err)

	err = validateCredentialValues("ohnoitsme", "ohnoitsme")
	assert.Equal(t, errSameUsernamePassword, err)

	err = validateCredentialValues("wowwow", "oh")
	assert.Equal(t, errInvalidPassword, err)

	err = validateCredentialValues("um", "ohasdf")
	assert.Equal(t, errInvalidUsername, err)

	err = validateCredentialValues("wow!!!!!!", "oasdfasdfh")
	assert.Equal(t, errInvalidUsername, err)

	err = validateCredentialValues("wowwow", "oasdfasdfh!!!!")
	assert.Equal(t, errInvalidPassword, err)
}
