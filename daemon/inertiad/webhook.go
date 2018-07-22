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

	payload, err := webhook.Parse(r, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
		return
	}

	switch event := payload.GetEventType(); event {
	case webhook.PushEvent:
		processPushEvent(payload, os.Stdout)
	// case webhook.PullEvent:
	// 	processPullRequestEvent(payload)
	default:
		fmt.Fprintln(os.Stdout, "Unrecognized event type")
	}
}

// specialized handler for docker webhooks
func dockerWebhookHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := webhook.ParseDocker(r, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
		return
	}

	fmt.Fprintln(os.Stdout, payload.GetPusher())
	fmt.Fprintln(os.Stdout, payload.GetTag())
	fmt.Fprintln(os.Stdout, payload.GetRepoName())
	fmt.Fprintln(os.Stdout, payload.GetName())
	fmt.Fprintln(os.Stdout, payload.GetOwner())
}

// processPushEvent prints information about the given PushEvent.
func processPushEvent(payload webhook.Payload, out io.Writer) {
	branch := common.GetBranchFromRef(payload.GetRef())

	fmt.Fprintln(out, "Received PushEvent")
	fmt.Fprintln(out, fmt.Sprintf("Repository Name: %s", payload.GetRepoName()))
	fmt.Fprintln(out, fmt.Sprintf("Repository Git URL: %s", payload.GetGitURL()))
	fmt.Fprintln(out, fmt.Sprintf("Branch: %s", branch))

	// Ignore event if repository not set up yet, otherwise
	// let deploy() handle the update.
	if deployment == nil {
		fmt.Fprintln(out, "No deployment detected - try running 'inertia $REMOTE up'")
		return
	}

	// Check for matching remotes
	err := deployment.CompareRemotes(payload.GetSSHURL())
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}

	// Check for matching branch
	if deployment.GetBranch() != branch {
		fmt.Fprintln(out, fmt.Sprintf("Event branch %s does not match deployed branch %s ignoring event", branch, deployment.GetBranch()))
		return
	}

	// If branches match, deploy
	fmt.Fprintln(out, fmt.Sprintf("Event branch matches deployed branch %s", branch))
	cli, err := docker.NewEnvClient()
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}
	defer cli.Close()

	// Deploy project
	err = deployment.Deploy(cli, os.Stdout, project.DeployOptions{
		SkipUpdate: false,
	})
	if err != nil {
		fmt.Fprintln(out, err.Error())
	}
}
