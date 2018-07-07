package main

import (
	"fmt"
	"io"
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
	outStream := os.Stdout

	payload, err := webhook.Parse(r, outStream)
	if err != nil {
		fmt.Fprintf(outStream, err.Error())
		return
	}

	switch event := payload.GetEventType(); event {
	case webhook.PushEvent:
		processPushEvent(payload, outStream)
	// case webhook.PullEvent:
	// 	processPullRequestEvent(payload)
	default:
		fmt.Fprintf(outStream, "Unrecognized event type")
	}
}

// processPushEvent prints information about the given PushEvent.
func processPushEvent(payload webhook.Payload, out io.Writer) {
	branch := common.GetBranchFromRef(payload.GetRef())

	fmt.Fprintf(out, "Received PushEvent")
	fmt.Fprintf(out, fmt.Sprintf("Repository Name: %s", payload.GetRepoName()))
	fmt.Fprintf(out, fmt.Sprintf("Repository Git URL: %s", payload.GetGitURL()))
	fmt.Fprintf(out, fmt.Sprintf("Branch: %s", branch))

	// Ignore event if repository not set up yet, otherwise
	// let deploy() handle the update.
	if deployment == nil {
		fmt.Fprintf(out, "No deployment detected - try running 'inertia $REMOTE up'")
		return
	}

	// Check for matching remotes
	err := deployment.CompareRemotes(payload.GetSSHURL())
	if err != nil {
		fmt.Fprintf(out, err.Error())
		return
	}

	// If branches match, deploy, otherwise ignore the event.
	if deployment.GetBranch() == branch {
		fmt.Fprintf(out, "Event branch matches deployed branch "+branch)
		cli, err := docker.NewEnvClient()
		if err != nil {
			fmt.Fprintf(out, err.Error())
			return
		}
		defer cli.Close()

		// Deploy project
		err = deployment.Deploy(cli, os.Stdout, project.DeployOptions{
			SkipUpdate: false,
		})
		if err != nil {
			fmt.Fprintf(out, err.Error())
		}
	} else {
		fmt.Fprintf(out,
			"Event branch "+branch+" does not match deployed branch "+
				deployment.GetBranch()+" - ignoring event.",
		)
	}
}
