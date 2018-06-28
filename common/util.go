package common

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// GetFullPath returns the absolute path of the config file.
func GetFullPath(relPath string) (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, relPath), nil
}

// GenerateRandomString creates a rand.Reader-generated
// string for use with simple secrets and identifiers
func GenerateRandomString() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CheckForDockerCompose returns false if current directory is a
// not a docker-compose project
func CheckForDockerCompose(cwd string) bool {
	dockerComposeYML := filepath.Join(cwd, "docker-compose.yml")
	_, err := os.Stat(dockerComposeYML)
	return !os.IsNotExist(err)
}

// CheckForDockerfile returns false if current directory is a
// not a Dockerfile project
func CheckForDockerfile(cwd string) bool {
	dockerfilePath := filepath.Join(cwd, "Dockerfile")
	_, err := os.Stat(dockerfilePath)
	return !os.IsNotExist(err)
}

// CheckForProcfile returns false if current directory is not a
// Heroku project
func CheckForProcfile(cwd string) bool {
	procfilePath := filepath.Join(cwd, "Procfile")
	_, err := os.Stat(procfilePath)
	return !os.IsNotExist(err)
}

// RemoveContents removes all files within given directory, returns nil if successful
func RemoveContents(directory string) error {
	d, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(directory, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseDate parses a date in format "2006-01-02T15:04:05.000Z"
func ParseDate(dateString string) *time.Time {
	layout := "2006-01-02T15:04:05.000Z"
	t, _ := time.Parse(layout, dateString)
	return &t
}

// ParseInt64 parses a string into an int64 value
func ParseInt64(value string) (int64, error) {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}
