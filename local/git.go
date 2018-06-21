package local

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetRepoRemote reads a remote URL
func GetRepoRemote(remote string) (string, error) {
	arg := "remote." + remote + ".url"
	out, err := exec.Command("git", "config", "--get", arg).CombinedOutput()
	if err != nil {
		return "", errors.New(err.Error() + ": " + string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// GetRepoCurrentBranch returns the current repository branch
func GetRepoCurrentBranch() (string, error) {
	out, err := exec.Command("git", "symbolic-ref", "--short", "HEAD").CombinedOutput()
	if err != nil {
		return "", errors.New(err.Error() + ": " + string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// checkForGit returns an error if we're not in a git repository.
func checkForGit(cwd string) error {
	// Quick failure if no .git folder.
	gitFolder := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitFolder); os.IsNotExist(err) {
		return errors.New("this does not appear to be a git repository")
	}

	// Also fail if no origin detected
	url, err := GetRepoRemote("origin")
	if err != nil {
		return err
	}
	if url == "" {
		return errors.New("no remote origin set")
	}

	return nil
}
