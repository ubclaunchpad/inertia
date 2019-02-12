package local

import (
	"errors"
	"os"
	"runtime"
)

const (
	// EnvSSHPassphrase is the key used to fetch PEM key passphrases
	EnvSSHPassphrase = "PEM_PASSPHRASE"
)

// GetHomePath returns the homepath based on the operating system used
func GetHomePath() (string, error) {
	if runtime.GOOS == "windows" {
		// Performs check for HOME env variable first
		if home := os.Getenv("HOME"); home != "" {
			return home, nil
		}

		// Performs check for USERPROFILE as backup
		if home := os.Getenv("USERPROFILE"); home != "" {
			return home, nil
		}

		// Builds HOME from HOMEDRIVE and HOMEPATH as default
		drive := os.Getenv("HOMEDRIVE")
		var path = os.Getenv("HOMEPATH")
		home := drive + path
		if drive == "" || path == "" {
			return "", errors.New("HOMEDRIVE, HOMEPATH, or USERPROFILE environment variables are blank")
		}

		return home, nil
	}
	// Unix system as default
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}
	return "", errors.New("HOME environment variable is blank")
}
