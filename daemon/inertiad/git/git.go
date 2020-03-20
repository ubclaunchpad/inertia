package git

import (
	"errors"
	"fmt"
	"io"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
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

// RepoOptions declares options for a project repository
type RepoOptions struct {
	Directory string
	Branch    string
	Auth      transport.AuthMethod
}

// InitializeRepository sets up a project repository for the first time
func InitializeRepository(remoteURL string, opts RepoOptions, w io.Writer) (*gogit.Repository, error) {
	fmt.Fprintln(w, "Setting up project...")
	repo, err := clone(remoteURL, opts, w)
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
func clone(remoteURL string, opts RepoOptions, out io.Writer) (*gogit.Repository, error) {
	// Preserve existing files by creating a repository, setting a remote, then
	// updating the directory
	repo, err := gogit.PlainInit(opts.Directory, false)
	if err != nil {
		return nil, fmt.Errorf("failed to init bare repository: %s", err.Error())
	}

	fmt.Fprintf(out, "Setting up repository from %s...\n", remoteURL)
	if _, err = repo.CreateRemote(&config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{remoteURL},
		Fetch: []config.RefSpec{"refs/*:refs/*"},
	}); err != nil {
		return nil, fmt.Errorf("failed to set remote: %s", err.Error())
	}

	// Fetch repository contents
	if err = UpdateRepository(repo, opts, out); err != nil {
		return nil, err
	}

	// Confirm if pull has completed.
	if _, err = repo.Head(); err != nil {
		return nil, err
	}

	return repo, nil
}

// UpdateRepository pulls and checkouts given branch from repository
func UpdateRepository(repo *gogit.Repository, opts RepoOptions, out io.Writer) error {
	tree, err := repo.Worktree()
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Fetching repository...")
	err = repo.Fetch(&gogit.FetchOptions{
		RemoteName: "origin",
		Auth:       opts.Auth,
		RefSpecs:   []config.RefSpec{"refs/*:refs/*"},
		Tags:       gogit.AllTags,
		Progress:   out,
		Force:      true,
	})
	if err = SimplifyGitErr(err); err != nil {
		return err
	}

	var ref = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", opts.Branch))
	fmt.Fprintf(out, "Checking out '%s'...\n", ref)
	err = tree.Checkout(&gogit.CheckoutOptions{
		Branch: ref,
		Force:  true,
	})
	if err = SimplifyGitErr(err); err != nil {
		return err
	}

	fmt.Fprintln(out, "Pulling from origin...")
	err = tree.Pull(&gogit.PullOptions{
		RemoteName:    "origin",
		ReferenceName: ref,
		Auth:          opts.Auth,
		Progress:      out,
		Force:         true,

		RecurseSubmodules: gogit.DefaultSubmoduleRecursionDepth,
	})
	return SimplifyGitErr(err)
}
