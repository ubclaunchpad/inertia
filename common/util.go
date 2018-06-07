package common

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GenerateRandomString creates a rand.Reader-generated
// string for use with simple secrets and identifiers
func GenerateRandomString() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CheckForDockerCompose returns error if current directory is a
// not a docker-compose project
func CheckForDockerCompose(cwd string) bool {
	dockerComposeYML := filepath.Join(cwd, "docker-compose.yml")
	dockerComposeYAML := filepath.Join(cwd, "docker-compose.yaml")
	_, err := os.Stat(dockerComposeYML)
	YMLnotPresent := os.IsNotExist(err)
	_, err = os.Stat(dockerComposeYAML)
	YAMLnotPresent := os.IsNotExist(err)
	return !(YMLnotPresent && YAMLnotPresent)
}

// CheckForDockerfile returns error if current directory is a
// not a Dockerfile project
func CheckForDockerfile(cwd string) bool {
	dockerfilePath := filepath.Join(cwd, "Dockerfile")
	_, err := os.Stat(dockerfilePath)
	dockerfileNotPresent := os.IsNotExist(err)
	return !dockerfileNotPresent
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

// ExtractRepository gets the project name from its URL in the form [username]/[project]
func ExtractRepository(URL string) string {
	const DefaultName = "$YOUR_REPOSITORY"
	re, err := regexp.Compile(":|/")
	if err != nil {
		return DefaultName
	}
	r := re.Split(strings.TrimSuffix(URL, ".git"), -1)
	return strings.Join(r[len(r)-2:], "/")
}
