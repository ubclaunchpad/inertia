package daemon

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	docker "github.com/docker/docker/client"
	"github.com/google/go-github/github"
	"github.com/ubclaunchpad/inertia/common"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// processPushEvent prints information about the given PushEvent.
func processPushEvent(event *github.PushEvent) {
	repo := event.GetRepo()
	branch := common.GetBranchFromRef(event.GetRef())
	println("Received PushEvent")
	println(fmt.Sprintf("Repository Name: %s", *repo.Name))
	println(fmt.Sprintf("Repository Git URL: %s", *repo.GitURL))
	println(fmt.Sprintf("Branch: %s", branch))

	// Ignore event if repository not set up yet, otherwise
	// let deploy() handle the update.
	err := common.CheckForGit(projectDirectory)
	if err != nil {
		println("No git repository present - try running 'inertia $REMOTE up'")
		return
	}

	localRepo, err := git.PlainOpen(projectDirectory)
	if err != nil {
		println(err)
		return
	}

	// Check for matching remotes
	err = common.CompareRemotes(localRepo, common.GetSSHRemoteURL(*repo.GitURL))
	if err != nil {
		println(err)
		return
	}

	// If branches match, deploy, otherwise ignore the event.
	head, err := localRepo.Head()
	if err != nil {
		println(err)
		return
	}
	if head.Name().Short() == branch {
		println("Event branch matches deployed branch " + head.Name().Short())
		cli, err := docker.NewEnvClient()
		if err != nil {
			println(err)
			return
		}
		defer cli.Close()
		err = deploy(localRepo, branch, cli, os.Stdout)
		if err != nil {
			println(err)
		}
	} else {
		println(
			"Event branch " + head.Name().Short() + " does not match deployed branch " +
				branch + " - ignoring event.",
		)
	}
}

// processPullRequestEvent prints information about the given PullRequestEvent.
// Handling PRs is unnecessary because merging one will trigger a PushEvent.
// For now, simply logs events - may in the future do something configured
// by the user.
func processPullRequestEvent(event *github.PullRequestEvent) {
	repo := event.GetRepo()
	pr := event.GetPullRequest()
	merged := "false"
	if *pr.Merged {
		merged = "true"
	}
	println("Received PullRequestEvent")
	println(fmt.Sprintf("Repository Name: %s", *repo.Name))
	println(fmt.Sprintf("Repository Git URL: %s", *repo.GitURL))
	println(fmt.Sprintf("Ref: %s", pr.GetBase().GetRef()))
	println(fmt.Sprintf("Merge status: %v", merged))
}

// setUpProject sets up a project for the first time
func setUpProject(remoteURL, branch string, w io.Writer) error {
	fmt.Fprintln(w, "Setting up project...")
	pemFile, err := os.Open(daemonGithubKeyLocation)
	if err != nil {
		return err
	}
	auth, err := getGithubKey(pemFile)
	if err != nil {
		return err
	}

	// Clone project
	_, err = common.Clone(projectDirectory, remoteURL, branch, auth, w)
	if err != nil {
		if err == common.ErrInvalidGitAuthentication {
			return gitAuthFailedErr(daemonGithubKeyLocation)
		}
		return err
	}
	return nil
}

// GetGithubKey returns an ssh.AuthMethod from the given io.Reader
// for use with the go-git library
func getGithubKey(pemFile io.Reader) (ssh.AuthMethod, error) {
	bytes, err := ioutil.ReadAll(pemFile)
	if err != nil {
		return nil, err
	}
	return ssh.NewPublicKeys("git", bytes, "")
}
