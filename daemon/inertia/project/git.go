package project

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertia/auth"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	// ErrInvalidGitAuthentication is returned when handshake with a git remote fails
	ErrInvalidGitAuthentication = errors.New("git authentication failed")
)

// SimplifyGitErr checks errors that involve git remote operations and simplifies them
// to ErrInvalidGitAuthentication if possible
func SimplifyGitErr(err error) error {
	if err != nil && err != git.NoErrAlreadyUpToDate {
		if err == transport.ErrInvalidAuthMethod || err == transport.ErrAuthorizationFailed || strings.Contains(err.Error(), "unable to authenticate") {
			return ErrInvalidGitAuthentication
		}
		return err
	}
	return nil
}

// initializeRepository sets up a project repository for the first time
func initializeRepository(remoteURL, branch string, authMethod ssh.AuthMethod, w io.Writer) (*git.Repository, error) {
	fmt.Fprintln(w, "Setting up project...")
	// Clone project
	repo, err := clone(Directory, remoteURL, branch, authMethod, w)
	if err != nil {
		if err == ErrInvalidGitAuthentication {
			return nil, auth.GitAuthFailedErr()
		}
		return nil, err
	}
	return repo, nil
}

// clone wraps git.PlainClone() and returns a more helpful error message
// if the given error is an authentication-related error.
func clone(directory, remoteURL, branch string, auth ssh.AuthMethod, out io.Writer) (*git.Repository, error) {
	fmt.Fprintf(out, "Cloning branch %s from %s...\n", branch, remoteURL)
	ref := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
	repo, err := git.PlainClone(directory, false, &git.CloneOptions{
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

// forcePull deletes the project directory and makes a fresh clone of given repo
// git.Worktree.Pull() only supports merges that can be resolved as a fast-forward
func forcePull(directory string, repo *git.Repository, auth ssh.AuthMethod, out io.Writer) (*git.Repository, error) {
	fmt.Fprintln(out, "Making a force pull...")
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	remoteURL := common.GetSSHRemoteURL(remotes[0].Config().URLs[0])
	branch := head.Name().Short()

	err = common.RemoveContents(directory)
	if err != nil {
		return nil, err
	}
	repo, err = clone(directory, remoteURL, branch, auth, out)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// updateRepository pulls and checkouts given branch from repository
func updateRepository(directory string, repo *git.Repository, branch string, auth ssh.AuthMethod, out io.Writer) error {
	tree, err := repo.Worktree()
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Fetching repository...")
	err = repo.Fetch(&git.FetchOptions{
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
	err = tree.Checkout(&git.CheckoutOptions{
		Branch: ref,
	})
	err = SimplifyGitErr(err)
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Pulling from origin...")
	err = tree.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: ref,
		Auth:          auth,
		Progress:      out,
	})
	err = SimplifyGitErr(err)
	if err != nil {
		if err == git.ErrForceNeeded {
			// If pull fails, attempt a force pull before returning error
			fmt.Fprintln(out, "Fast-forward failed - a force pull is required.")
			_, err := forcePull(directory, repo, auth, out)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
