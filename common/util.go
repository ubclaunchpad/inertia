package common

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
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

// FlushRoutine continuously writes everything in given ReadCloser
// to a ResponseWriter. Use this as a goroutine.
func FlushRoutine(w io.Writer, rc io.ReadCloser, stop chan struct{}) {
	buffer := make([]byte, 100)
ROUTINE:
	for {
		select {
		case <-stop:
			Flush(w, rc, buffer)
			break ROUTINE
		default:
			// Read from pipe then write to ResponseWriter and flush it,
			// sending the copied content to the client.
			err := Flush(w, rc, buffer)
			if err != nil {
				break ROUTINE
			}
		}
	}
}

// Flush emptires reader into buffer and flushes it to writer
func Flush(w io.Writer, rc io.ReadCloser, buffer []byte) error {
	n, err := rc.Read(buffer)
	if err != nil {
		rc.Close()
		return err
	}
	data := buffer[0:n]
	w.Write(data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Clear the buffer.
	for i := 0; i < n; i++ {
		buffer[i] = 0
	}
	return nil
}

// BuildTar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)
// Sourced from https://gist.github.com/sdomino/e6bc0c98f87843bc26bb#file-targz-go
func BuildTar(dir string, outputs ...io.Writer) error {

	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	mw := io.MultiWriter(outputs...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(dir, func(file string, fi os.FileInfo, err error) error {
		// return on any error
		if err != nil {
			return err
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, dir, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// return on non-regular files
		if !fi.Mode().IsRegular() {
			return nil
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		// copy file data into tar writer
		_, err = io.Copy(tw, f)
		return err
	})
}

// ExtractRepository gets the project name from its URL in the form [username]/[project]
func ExtractRepository(URL string) (string, error) {
	r, err := regexp.Compile(`.com(/|:)(\w+/\w+)`)
	if err != nil {
		return "", err
	}

	remoteString := r.FindStringSubmatch(URL)
	if len(remoteString) != 3 {
		return "", fmt.Errorf("Failed to extract repository name with remote url %s", URL)
	}
	return remoteString[2], nil
}
