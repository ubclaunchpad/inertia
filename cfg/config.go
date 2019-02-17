package cfg

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

// Write writes configuration to Inertia config file at path. Optionally
// takes io.Writers.
func Write(path string, data interface{}, writers ...io.Writer) error {
	if len(writers) == 0 && path == "" {
		return errors.New("nothing to write to")
	}

	// If io.Writers are given, attach all writers
	var writer io.Writer
	if len(writers) > 0 {
		writer = io.MultiWriter(writers...)
	}

	// If path is given, attach file writer
	if path != "" {
		w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}

		// Overwrite file if file exists
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			ioutil.WriteFile(path, []byte(""), 0644)
		} else if err != nil {
			return err
		}

		// Set writer
		if writer != nil {
			writer = io.MultiWriter(writer, w)
		} else {
			writer = w
		}
	} else {
		writer = os.Stdout
	}

	// Write configuration to writers
	encoder := toml.NewEncoder(writer)
	return encoder.Encode(data)
}
