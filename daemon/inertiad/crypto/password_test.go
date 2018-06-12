package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	unhashed := "amazing"
	hashed, err := HashPassword(unhashed)
	assert.Nil(t, err)
	assert.NotEqual(t, unhashed, hashed)
}

func TestCorrectPassword(t *testing.T) {
	unhashed := "amazing"
	hashed, err := HashPassword(unhashed)
	assert.Nil(t, err)
	assert.NotEqual(t, unhashed, hashed)

	correct := CorrectPassword(hashed, unhashed)
	assert.True(t, correct)

	correct = CorrectPassword(hashed, "ummmmm")
	assert.False(t, correct)
}

func TestValidateCredentialValues(t *testing.T) {
	err := ValidateCredentialValues("finasdfsdfe", "okaasdfasdy")
	assert.Nil(t, err)

	err = ValidateCredentialValues("0123456789a", "0123456789")
	assert.Nil(t, err)

	err = ValidateCredentialValues("ohnoitsme", "ohnoitsme")
	assert.Equal(t, errSameUsernamePassword, err)

	err = ValidateCredentialValues("wowwow", "oh")
	assert.Equal(t, errInvalidPassword, err)

	err = ValidateCredentialValues("um", "ohasdf")
	assert.Equal(t, errInvalidUsername, err)

	err = ValidateCredentialValues("wow!!!!!!", "oasdfasdfh")
	assert.Equal(t, errInvalidUsername, err)

	err = ValidateCredentialValues("wowwow", "oasdfasdfh!!!!")
	assert.Equal(t, errInvalidPassword, err)
}
