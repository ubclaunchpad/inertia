package common

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

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

// Extract gets the project name from its URL in the form [username]/[project]
func Extract(URL string) string {
	r, _ := regexp.Compile(".com(/|:)(\\w+/\\w+)")
	return r.FindStringSubmatch(URL)[2]
}
