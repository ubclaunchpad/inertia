package git

import (
	"errors"
	"fmt"
	"io"
	"strings"

	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
)

var (
	// ErrInvalidGitAuthentication is returned when handshake with a git remote fails
	ErrInvalidGitAuthentication = errors.New("git authentication failed")
)

// SimplifyGitErr checks errors that involve git remote operations and simplifies them
// to ErrInvalidGitAuthentication if possible
func SimplifyGitErr(err error) error {
	if err != nil && err != gogit.NoErrAlreadyUpToDate {
		if err == transport.ErrInvalidAuthMethod || err == transport.ErrAuthorizationFailed || strings.Contains(err.Error(), "unable to authenticate") {
			return ErrInvalidGitAuthentication
		}
		return err
	}
	return nil
}

// InitializeRepository sets up a project repository for the first time
func InitializeRepository(directory, remoteURL, branch string, authMethod transport.AuthMethod, w io.Writer) (*gogit.Repository, error) {
	fmt.Fprintln(w, "Setting up project...")
	// Clone project
	repo, err := clone(directory, remoteURL, branch, authMethod, w)
	if err != nil {
		if err == ErrInvalidGitAuthentication {
			return nil, AuthFailedErr()
		}
		return nil, err
	}
	return repo, nil
}

// clone wraps gogit.PlainClone() and returns a more helpful error message
// if the given error is an authentication-related error.
func clone(directory, remoteURL, branch string, auth transport.AuthMethod, out io.Writer) (*gogit.Repository, error) {
	fmt.Fprintf(out, "Cloning branch %s from %s...\n", branch, remoteURL)
	ref := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
	repo, err := gogit.PlainClone(directory, false, &gogit.CloneOptions{
		URL:           remoteURL,
		Auth:          auth,
		Progress:      out,
		ReferenceName: ref,
	})
	err = SimplifyGitErr(err)
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

// UpdateRepository pulls and checkouts given branch from repository
func UpdateRepository(directory string, repo *gogit.Repository, branch string, auth transport.AuthMethod, out io.Writer) error {
	tree, err := repo.Worktree()
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Fetching repository...")
	err = repo.Fetch(&gogit.FetchOptions{
		Auth:     auth,
		RefSpecs: []config.RefSpec{"refs/*:refs/*"},
		Progress: out,
	})
	err = SimplifyGitErr(err)
	if err != nil {
		return err
	}

	ref := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
	fmt.Fprintf(out, "Checking out %s...\n", ref)
	err = tree.Checkout(&gogit.CheckoutOptions{
		Branch: ref,
	})
	err = SimplifyGitErr(err)
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Pulling from origin...")
	err = tree.Pull(&gogit.PullOptions{
		RemoteName:    "origin",
		ReferenceName: ref,
		Auth:          auth,
		Progress:      out,
		Force:         true,
	})
	return SimplifyGitErr(err)
}
