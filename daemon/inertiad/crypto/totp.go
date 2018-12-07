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

// GenerateSecretKey creates a new key which can be turned into string or image
func GenerateSecretKey(accountName string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuerName,
		AccountName: accountName,
		Period:      totpPeriod,
		SecretSize:  totpSecretSize,
		Digits:      totpDigits,
		Algorithm:   totpAlgorithm,
	})
}

// ValidatePasscode validates one-time passcode against original secret key
func ValidatePasscode(passcode string, secret string) bool {
	return totp.Validate(passcode, secret)
}

// GenerateBackupCodes generates an array of backup code strings in
// Github format.
//
// Example:
// b2e03-ffbcf
// cebe6-b1bdd
// ...
func GenerateBackupCodes() []string {
	backupCodes := make([]string, totpNoBackupCodes)
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
