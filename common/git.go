package common

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
)

// GetLocalRepo gets the repo from disk.
func GetLocalRepo() (*git.Repository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return git.PlainOpen(cwd)
}

// CheckForGit returns an error if we're not in a git repository.
func CheckForGit(cwd string) error {
	// Quick failure if no .git folder.
	gitFolder := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitFolder); os.IsNotExist(err) {
		return errors.New("this does not appear to be a git repository")
	}

	repo, err := git.PlainOpen(cwd)
	if err != nil {
		return err
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return err
	}

	// Also fail if no remotes detected.
	if len(remotes) == 0 {
		return errors.New("there are no remotes associated with this repository")
	}

	return nil
}

// GetSSHRemoteURL gets the URL of the given remote in the form
// "git@github.com:[USER]/[REPOSITORY].git"
func GetSSHRemoteURL(url string) string {
	re, _ := regexp.Compile("(https://)|(git://)")

	sshURL := re.ReplaceAllString(url, "git@")
	if sshURL != url {
		sshURL = strings.Replace(sshURL, "/", ":", 1)
	}

	// special bitbucket https case?
	lastIndex := strings.LastIndex(sshURL, "@")
	if lastIndex > 3 {
		sshURL = "git@" + sshURL[lastIndex+1:]
	}

	if !strings.HasSuffix(sshURL, ".git") {
		sshURL = sshURL + ".git"
	}
	return sshURL
}

// GetBranchFromRef gets the branch name from a git ref of form refs/...
func GetBranchFromRef(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}
