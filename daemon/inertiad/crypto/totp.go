package crypto

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	totpIssuerName       = "Inertia"
	totpPeriod           = 30
	totpSecretSize       = 10
	totpDigits           = 6
	totpAlgorithm        = otp.AlgorithmSHA1
	totpNoBackupCodes    = 10
	totpBackupCodeLength = 5
)

// Generates secret key object which can be turned into string or image
func generateSecretKey(accountName string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuerName,
		AccountName: accountName,
		Period:      totpPeriod,
		SecretSize:  totpSecretSize,
		Digits:      totpDigits,
		Algorithm:   totpAlgorithm,
	})
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
func generateBackupCodes() (backupCodes [totpNoBackupCodes]string) {
	for i := 0; i < totpNoBackupCodes; i++ {
		// get random bytes
		randomBytes := make([]byte, totpBackupCodeLength)
		rand.Read(randomBytes)

		// convert to hex string
		codeHex := hex.EncodeToString(randomBytes)

		// split with dash
		code := codeHex[:5] + "-" + codeHex[5:]
		backupCodes[i] = code
	}
	return backupCodes
}
