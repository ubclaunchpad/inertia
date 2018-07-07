package main

import (
	"fmt"
	"net/http"
	"os"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/webhook"
)

var webhookSecret = "inertia"

// webhookHandler writes a response to a request into the given ResponseWriter.
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, common.MsgDaemonOK)

	payload, err := webhook.Parse(r)
	if err != nil {
		println(err)
		return
	}

	switch event := payload.GetEventType(); event {
	case webhook.PushEvent:
		processPushEvent(payload)
	// case webhook.PullEvent:
	// 	processPullRequestEvent(payload)
	default:
		println("Unrecognized event type")
	}
}

// processPushEvent prints information about the given PushEvent.
func processPushEvent(payload webhook.Payload) {
	branch := common.GetBranchFromRef(payload.GetRef())

	println("Received PushEvent")
	println(fmt.Sprintf("Repository Name: %s", payload.GetRepoName()))
	println(fmt.Sprintf("Repository Git URL: %s", payload.GetGitURL()))
	println(fmt.Sprintf("Branch: %s", branch))

	// Ignore event if repository not set up yet, otherwise
	// let deploy() handle the update.
	if deployment == nil {
		println("No deployment detected - try running 'inertia $REMOTE up'")
		return
	}

	// Check for matching remotes
	err := deployment.CompareRemotes(payload.GetSSHURL())
	if err != nil {
		println(err)
		return
	}

	// If branches match, deploy, otherwise ignore the event.
	if deployment.GetBranch() == branch {
		println("Event branch matches deployed branch " + branch)
		cli, err := docker.NewEnvClient()
		if err != nil {
			println(err)
			return
		}
		defer cli.Close()

		// Deploy project
		err = deployment.Deploy(cli, os.Stdout, project.DeployOptions{
			SkipUpdate: false,
		})
		if err != nil {
			println(err)
		}
	} else {
		println(
			"Event branch " + branch + " does not match deployed branch " +
				deployment.GetBranch() + " - ignoring event.",
		)
	}
}
