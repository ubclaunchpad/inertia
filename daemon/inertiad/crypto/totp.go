package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	TotpIssuerName       = "Inertia"
	TotpPeriod           = 30
	TotpSecretSize       = 10
	TotpDigits           = 6
	TotpAlgorithm        = otp.AlgorithmSHA1
	TotpNoBackupCodes    = 10
	TotpBackupCodeLength = 5
)

// Generates secret key object which can be turned into string or image
func generateSecretKey(accountName string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      TotpIssuerName,
		AccountName: accountName,
		Period:      TotpPeriod,
		SecretSize:  TotpSecretSize,
		Digits:      TotpDigits,
		Algorithm:   TotpAlgorithm,
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

// Validate one-time passcode against original secret key
func validatePasscode(passcode string, secret string) bool {
	return totp.Validate(passcode, secret)
}

// Generates backup code strings in Github format
//
// b2e03-ffbcf
// cebe6-b1bdd
// ...
func generateBackupCodes() (backupCodes [TotpNoBackupCodes]string) {
	for i := 0; i < TotpNoBackupCodes; i++ {
		// get random bytes
		randomBytes := make([]byte, TotpBackupCodeLength)
		rand.Read(randomBytes)

		// convert to hex string
		codeHex := hex.EncodeToString(randomBytes)

		// split with dash
		code := codeHex[:5] + "-" + codeHex[5:]
		backupCodes[i] = code
	}
	return backupCodes
}
