/*

Inertia-publish is a tool for helping publish Inertia releases to various distribution
channels.

For example, to generate completions for zsh:

	inertia contrib publish [channel]

Learn more about `inertia/contrib` tools:

	inertia contrib -h

*/
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
	templatedata "github.com/ubclaunchpad/inertia/contrib/inertia-publish/templates"
)

//go:generate go run github.com/UnnoTed/fileb0x b0x.yml

// TemplateVariables carries variables available to templates
type TemplateVariables struct {
	Version string
	Sha256  map[string]string
}

func generateSums(buildDir string) map[string]string {
	sums := make(map[string]string)
	if err := filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
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
		sums[build] = hex.EncodeToString(sum[:])
		return nil
	}); err != nil {
		out.Fatal(err)
	}

	// if no sums were generated, some setup is likely missing
	if len(sums) == 0 {
		out.Fatal("no binary sums generated - was '.scripts/build_release.sh' run?")
	}

	return sums
}

func newCommand(dir, name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd
}

func getCloneURL(repository string) string {
	// GIT_CREDENTIALS should be $USER:$TOKEN
	if creds := os.Getenv("GIT_CREDENTIALS"); creds != "" {
		println("Using GIT_CREDENTIALS to clone")
		// "Personal access tokens can only be used for HTTPS Git operations.""
		// https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token#using-a-token-on-the-command-line
		return fmt.Sprintf("https://%s@github.com/%s.git", creds, repository)
	}
	return fmt.Sprintf("https://github.com/%s.git", repository)
}

type channelConfig struct {
	Repository string
	Template   string
	PublishAs  string
}

func (c *channelConfig) Publish(version, outdir string, renderedTemplate []byte, dryRun bool) error {
	// load repository
	repoPath := filepath.Join(outdir, c.Repository)
	os.RemoveAll(repoPath)
	clone := newCommand(outdir, "git", "clone", getCloneURL(c.Repository),
		// clone into subdirectory
		c.Repository)
	if err := clone.Run(); err != nil {
		return err
	}

	// update file
	publishFile := filepath.Join(repoPath, c.PublishAs)
	os.Remove(publishFile)
	if err := ioutil.WriteFile(publishFile, renderedTemplate, os.ModePerm); err != nil {
		return err
	}

	// commit changes
	commit := newCommand(repoPath, "git", "commit", "-a",
		"-m", fmt.Sprintf("inertia-publish: release %s", version))
	if err := commit.Run(); err != nil {
		return err
	}

	// publish
	if dryRun {
		out.Println("dryrun enabled - aborting")
		return nil
	}
	publish := newCommand(repoPath, "git", "push", "origin", "HEAD")
	return publish.Run()
}

var channelConfigs = map[string]channelConfig{
	"homebrew": {
		Repository: "ubclaunchpad/homebrew-tap",
		Template:   "homebrew.rb",
		PublishAs:  "inertia.rb",
	},
	"scoop": {
		Repository: "ubclaunchpad/scoop-bucket",
		Template:   "scoop.json",
		PublishAs:  "inertia.json",
	},
}

func main() {
	var (
		flagOutput = "output"
		flagDryRun = "dryrun"
	)
	var publish = &cobra.Command{
		Use:    "inertia-publish [version] [channels]",
		Hidden: true,
		Args:   cobra.MinimumNArgs(1),
		Short:  "Publish Inertia releases to various distribution channels.",
		Run: func(cmd *cobra.Command, args []string) {
			outdir, _ := cmd.Flags().GetString(flagOutput)
			dryRun, _ := cmd.Flags().GetBool(flagDryRun)
			os.MkdirAll(outdir, os.ModePerm)
			data := TemplateVariables{
				Version: args[0],
				Sha256:  generateSums(outdir),
			}
			for _, channel := range args[1:] {
				conf, exists := channelConfigs[channel]
				if !exists {
					out.Fatalf("unknown channel %q", channel)
				}
				out.Printf("Preparing %q publish...\n", channel)

				rawTemplate, err := templatedata.ReadFile(filepath.Join("templates", conf.Template))
				if err != nil {
					out.Fatal(err)
				}
				tmpl, err := template.New(channel).Parse(string(rawTemplate))
				if err != nil {
					out.Fatal(err)
				}
				var rendered bytes.Buffer
				if err := tmpl.Execute(&rendered, data); err != nil {
					out.Fatal(err)
				}
				if err := conf.Publish(data.Version, outdir, rendered.Bytes(), dryRun); err != nil {
					out.Fatal(err)
				}
			}
		},
	}
	publish.Flags().String(flagOutput, "dist", "output directory for rendered templates and binaries")
	publish.Flags().Bool(flagDryRun, true, "disable publish step")
	if err := publish.Execute(); err != nil {
		out.Fatal(err)
	}
}
