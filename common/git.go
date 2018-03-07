package common

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	// ErrInvalidGitAuthentication is returned when handshake with a git remote fails
	ErrInvalidGitAuthentication = errors.New("git authentication failed")
)

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

// GetLocalRepo gets the repo from disk.
func GetLocalRepo() (*git.Repository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return git.PlainOpen(cwd)
}

// GetSSHRemoteURL gets the URL of the given remote in the form
// "git@github.com:[USER]/[REPOSITORY].git"
func GetSSHRemoteURL(url string) string {
	return strings.Replace(url, "https://github.com/", "git@github.com:", -1)
}

// Clone wraps git.PlainClone() and returns a more helpful error message
// if the given error is an authentication-related error.
func Clone(directory, remoteURL, branch string, auth ssh.AuthMethod, out io.Writer) (*git.Repository, error) {
	fmt.Fprintf(out, "Cloning branch %s from %s...\n", branch, remoteURL)
	ref := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
	repo, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:           remoteURL,
		Auth:          auth,
		Depth:         2,
		Progress:      out,
		ReferenceName: ref,
	})
	err = CheckGitRemoteErr(err)
	if err != nil {
		return nil, err
	}

	// Use this to confirm if pull has completed.
	_, err = repo.Head()
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// ForcePull deletes the project directory and makes a fresh clone of given repo
// git.Worktree.Pull() only supports merges that can be resolved as a fast-forward
func ForcePull(directory string, repo *git.Repository, auth ssh.AuthMethod, out io.Writer) (*git.Repository, error) {
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}
	remoteURL := GetSSHRemoteURL(remotes[0].Config().URLs[0])
	err = RemoveContents(directory)
	if err != nil {
		return nil, err
	}
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	repo, err = Clone(directory, remoteURL, head.Name().Short(), auth, out)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// UpdateRepository pulls and checkouts given branch from repository
func UpdateRepository(directory string, repo *git.Repository, branch string, auth ssh.AuthMethod, out io.Writer) error {
	tree, err := repo.Worktree()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Checking out and updating branch %s...\n", branch)
	ref := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
	err = tree.Pull(&git.PullOptions{
		Auth:          auth,
		Depth:         2,
		Progress:      out,
		ReferenceName: ref,
	})
	err = CheckGitRemoteErr(err)
	if err != nil {
		if err == git.ErrForceNeeded {
			// If pull fails, attempt a force pull before returning error
			fmt.Fprint(out, "Force pull required - making a fresh clone...")
			_, err := ForcePull(directory, repo, auth, out)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Checkout the pulled branch
	return tree.Checkout(&git.CheckoutOptions{Branch: ref})
}

// CompareRemotes checks if the given remote matches the remote of the given repository
func CompareRemotes(localRepo *git.Repository, remoteURL string) error {
	remotes, err := localRepo.Remotes()
	if err != nil {
		return err
	}
	localRemoteURL := GetSSHRemoteURL(remotes[0].Config().URLs[0])
	if localRemoteURL != remoteURL {
		return errors.New("The given remote URL does not match that of the repository in\nyour remote - try 'inertia [REMOTE] reset'")
	}
	return nil
}

// CheckGitRemoteErr checks errors that involve git remote operations and simplifies them
// to ErrInvalidGitAuthentication if possible
func CheckGitRemoteErr(err error) error {
	if err != nil && err != git.NoErrAlreadyUpToDate {
		if err == transport.ErrInvalidAuthMethod || err == transport.ErrAuthorizationFailed || strings.Contains(err.Error(), "unable to authenticate") {
			return ErrInvalidGitAuthentication
		}
		return err
	}
	return nil
}
