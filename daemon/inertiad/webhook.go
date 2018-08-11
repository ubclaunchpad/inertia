package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/webhook"
)

var webhookSecret = "inertia"

// webhookHandler receives and parses Git-based webhooks
// Supported vendors: Github, Gitlab, Bitbucket
// Supported events: push
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, common.MsgDaemonOK)

	payload, err := webhook.Parse(r)
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
	p, err := webhook.ParseDocker(r)
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
		return
	}

	fmt.Fprintf(os.Stdout, "Docker webhook event: %s:%s\n", p.GetRepoName(), p.GetTag())
}

// processPushEvent prints information about the given PushEvent.
func processPushEvent(p webhook.Payload, out io.Writer) {
	fmt.Fprintf(out, "%s push event: %s (%s)\n",
		p.GetSource(), p.GetRepoName(), p.GetRef())

	cli, err := containers.NewDockerClient()
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}
	defer cli.Close()

	// Ignore event if repository not set up yet, otherwise
	// let deploy() handle the update.
	if status, _ := deployment.GetStatus(cli); status.CommitHash == "" {
		fmt.Fprintln(out, msgNoDeployment)
		return
	}

	// Check for matching remotes
	err = deployment.CompareRemotes(p.GetSSHURL())
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}

	// Check for matching branch
	branch := common.GetBranchFromRef(p.GetRef())
	if deployment.GetBranch() != branch {
		fmt.Fprintf(out, "Ignoring event: event branch %s does not match deployed branch %s",
			branch, deployment.GetBranch())
		return
	}

	// If branches match, deploy
	fmt.Fprintf(out, "Accepting event: event branch %s matches deployed branch %s",
		branch, deployment.GetBranch())
	deploy, err := deployment.Deploy(cli, os.Stdout, project.DeployOptions{
		SkipUpdate: false,
	})
	if err != nil {
		fmt.Fprintln(out, err.Error())
	}

	err = deploy()
	if err != nil {
		fmt.Fprintln(out, err.Error())
	}
}
