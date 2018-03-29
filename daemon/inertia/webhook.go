package main

import (
	"fmt"
	"net/http"
	"os"

	docker "github.com/docker/docker/client"
	"github.com/google/go-github/github"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertia/project"
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
	err := common.CheckForGit(project.Directory)
	if err != nil {
		println("No git repository present - try running 'inertia $REMOTE up'")
		return
	}

	// Check for matching remotes
	err = deployment.CompareRemotes(common.GetSSHRemoteURL(repo.GetGitURL()))
	if err != nil {
		println(err.Error())
		return
	}

	// If branches match, deploy, otherwise ignore the event.
	if deployment.Branch == branch {
		println("Event branch matches deployed branch " + branch)
		cli, err := docker.NewEnvClient()
		if err != nil {
			println(err.Error())
			return
		}
		defer cli.Close()

		// Deploy project
		err = deployment.Deploy(false, cli, os.Stdout)
		if err != nil {
			println(err.Error())
		}
	} else {
		println(
			"Event branch " + branch + " does not match deployed branch " +
				deployment.Branch + " - ignoring event.",
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
