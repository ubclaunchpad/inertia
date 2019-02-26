package crypto

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCredentialFormatError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"not credential error", args{errors.New("robert")}, false},
		{"is credential error", args{errInvalidPassword}, true},
		{"is credential error", args{errInvalidUsername}, true},
		{"is credential error", args{errSameUsernamePassword}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCredentialFormatError(tt.args.err); got != tt.want {
				t.Errorf("IsCredentialFormatError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	unhashed := "amazing"
	hashed, err := HashPassword(unhashed)
	assert.NoError(t, err)
	assert.NotEqual(t, unhashed, hashed)
}

func TestCorrectPassword(t *testing.T) {
	unhashed := "amazing"
	hashed, err := HashPassword(unhashed)
	assert.NoError(t, err)
	assert.NotEqual(t, unhashed, hashed)

	correct := CorrectPassword(hashed, unhashed)
	assert.True(t, correct)

	correct = CorrectPassword(hashed, "ummmmm")
	assert.False(t, correct)
}

func TestValidateCredentialValues(t *testing.T) {
	err := ValidateCredentialValues("finasdfsdfe", "okaasdfasdy")
	assert.NoError(t, err)

	err = ValidateCredentialValues("0123456789a", "0123456789")
	assert.NoError(t, err)

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
