package daemon

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	docker "github.com/docker/docker/client"
	"github.com/google/go-github/github"
	"github.com/ubclaunchpad/inertia/common"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// gitHubWebHookHandler writes a response to a request into the given ResponseWriter.
func gitHubWebHookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, common.DaemonOkResp)

	payload, err := github.ValidatePayload(r, []byte(defaultSecret))
	if err != nil {
		println(err.Error())
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		println(err.Error())
		return
	}

	switch event := event.(type) {
	case *github.PushEvent:
		processPushEvent(event)
	case *github.PullRequestEvent:
		processPullRequestEvent(event)
	default:
		println("Unrecognized event type")
	}
}

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

	// Check for matching remotes
	localRepo, err := git.PlainOpen(projectDirectory)
	if err != nil {
		println(err.Error())
		return
	}
	err = common.CompareRemotes(localRepo, common.GetSSHRemoteURL(repo.GetGitURL()))
	if err != nil {
		println(err.Error())
		return
	}

	// If branches match, deploy, otherwise ignore the event.
	head, err := localRepo.Head()
	if err != nil {
		println(err.Error())
		return
	}
	if head.Name().Short() == branch {
		println("Event branch matches deployed branch " + branch)
		cli, err := docker.NewEnvClient()
		if err != nil {
			println(err.Error())
			return
		}
		defer cli.Close()
		err = deploy(localRepo, branch, projectName, cli, os.Stdout)
		if err != nil {
			println(err.Error())
		}
	} else {
		println(
			"Event branch " + branch + " does not match deployed branch " +
				head.Name().Short() + " - ignoring event.",
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
