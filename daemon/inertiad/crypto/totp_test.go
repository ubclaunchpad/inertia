package crypto

import (
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGeneration(t *testing.T) {
	key, err := generateSecretKey("TestAccountName")
	assert.Nil(t, err)

	// string form of secret
	secret := key.Secret()
	assert.Equal(t, len(secret), 16)

	// img form of secret
	_, err = key.Image(400, 400)
	assert.Nil(t, err)
}

func TestVerification(t *testing.T) {
	// enroll new user
	key, err := generateSecretKey("TestAccountName")
	assert.Nil(t, err)

	// get current time
	currentTime := time.Now()

	// generate current TOTP (valid)
	code, err := totp.GenerateCode(key.Secret(), currentTime)
	assert.Nil(t, err)
	assert.True(t, validatePasscode(code, key.Secret()))

	// generate TOTP before current window (invalid)
	badTime := currentTime.Add(time.Duration(-(TotpPeriod * 2) * time.Second))
	code, err = totp.GenerateCode(key.Secret(), badTime)
	assert.False(t, validatePasscode(code, key.Secret()))

	// generate TOTP after current window (invalid)
	badTime = currentTime.Add(time.Duration((TotpPeriod * 2) * time.Second))
	code, err = totp.GenerateCode(key.Secret(), badTime)
	assert.False(t, validatePasscode(code, key.Secret()))
}

func TestBackupCodes(t *testing.T) {
	codes := generateBackupCodes()
	assert.Equal(t, len(codes), TotpNoBackupCodes)
	for _, code := range codes {
		assert.Equal(t, len(code), 11)
	}
}