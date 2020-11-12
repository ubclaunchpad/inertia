package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
)

// channelConfig defines a distribution channel for Inertia binaries.
type channelConfig struct {
	// Repository is a GitHub repository that hosts this distribution channel, for example
	// `ubclaunchpad/homebrew-tap`.
	Repository string

	// Template should provide a template file used to define Inertia in this distribution
	// channel.
	//
	// Templates should be provided in the `templates` subdirectory, which will are compiled
	// into the publishing tool when you run `go generate ./...`. Templates can leverage
	// variables provided by the `TemplateVariables` struct.
	Template string

	// PublishAs defines the name of the file which the given `Template` should be saved
	// and committed as.
	PublishAs string
}

func (c *channelConfig) Publish(version, outdir string, renderedTemplate []byte, dryRun bool) error {
	// load repository
	repoPath := filepath.Join(outdir, c.Repository)
	os.RemoveAll(repoPath)
	clone := newCommand(outdir, "git", "clone", c.getCloneURL(),
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

// getCloneURL generates a URL that can be used to clone and push to this channel's
// repository.
func (c *channelConfig) getCloneURL() string {
	// GIT_CREDENTIALS should be $USER:$TOKEN
	if creds := os.Getenv("GIT_CREDENTIALS"); creds != "" {
		println("Using GIT_CREDENTIALS to clone")
		// "Personal access tokens can only be used for HTTPS Git operations.""
		// https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token#using-a-token-on-the-command-line
		return fmt.Sprintf("https://%s@github.com/%s.git", creds, c.Repository)
	}
	return fmt.Sprintf("https://github.com/%s.git", c.Repository)
}

// newCommand initializes a shell command
func newCommand(dir, name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd
}
