package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
)

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
