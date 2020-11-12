package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// TemplateVariables carries variables available to templates
type TemplateVariables struct {
	// Version of this release of Inertia.
	//
	// Usage in templates:
	//
	//     {{ .Version }}
	//
	// Version corresponds to tagged releases (https://github.com/ubclaunchpad/inertia/releases)
	// or placeholder values like `test`.
	Version string

	// Map of asset name to their corresponding Sha256 sum.
	//
	// Usage in templates:
	//
	//     {{ index .Sha256 "darwin.amd64" }}
	//
	// Assets created are defined in `.scripts/build_release_cli.sh`.
	Sha256 map[string]string
}

// newTemplateVariables initializes data available for templates.
//
// `assetsDir` should be a directory holding Inertia release assets, such as `inertia.v0.7.0.darwin.amd64`.
func newTemplateVariables(version, assetsDir string) (TemplateVariables, error) {
	// generate sums for release assets in the provided assetsDir
	sha256sums := make(map[string]string)
	if err := filepath.Walk(assetsDir, func(path string, info os.FileInfo, err error) error {
		parts := strings.Split(info.Name(), ".")
		if info.IsDir() || parts[0] != "inertia" {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		build := strings.Join(parts[len(parts)-2:], ".")
		sha256sums[build] = hex.EncodeToString(sum[:])
		return nil
	}); err != nil {
		return TemplateVariables{}, err
	}
	// if no sums were generated, some setup is likely missing
	if len(sha256sums) == 0 {
		return TemplateVariables{}, errors.New("no binary sums generated - was '.scripts/build_release.sh' run?")
	}

	return TemplateVariables{
		Version: version,
		Sha256:  sha256sums,
	}, nil
}
