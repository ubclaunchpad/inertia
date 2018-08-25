package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/webhook"
)

var webhookSecret = ""

// webhookHandler receives and parses Git-based webhooks
// Supported vendors: Github, Gitlab, Bitbucket
// Supported events: push
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// read
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := "unable to read payload: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		println(msg)
		return
	}

	// check type
	host, event := webhook.Type(r.Header)

	// ensure validity
	if webhookSecret == "" {
		println("warning: no webhook secret is set up yet! set one in inertia.toml and run inertia [remote] up")
	}
	if err := webhook.Verify(host, webhookSecret, r.Header, body); err != nil {
		msg := "unable to verify payload: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		println(msg)
		return
	}

	// retrieve payload
	payload, err := webhook.Parse(host, event, r.Header, body)
	if err != nil {
		msg := "unable to parse payload: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		println(msg)
		return
	}

	// process event
	switch event := payload.GetEventType(); event {
	case webhook.PushEvent:
		fmt.Fprint(w, common.MsgDaemonOK)
		processPushEvent(payload, os.Stdout)
	// case webhook.PullEvent:
	//	fmt.Fprint(w, common.MsgDaemonOK)
	// 	processPullRequestEvent(payload)
	default:
		http.Error(w, "unrecognized event type", http.StatusBadRequest)
		println("unrecognized event type")
	}
}

// specialized handler for docker webhooks
func dockerWebhookHandler(w http.ResponseWriter, r *http.Request) {
	p, err := webhook.ParseDocker(r)
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
		return
	}

	fmt.Printf("Received dockerhub webhook event: %s:%s\n", p.GetRepoName(), p.GetTag())
}

// processPushEvent prints information about the given PushEvent.
func processPushEvent(p webhook.Payload, out io.Writer) {
	fmt.Fprintf(out, "Received %s push event: %s (%s)\n",
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
		fmt.Fprintf(out, "Ignoring event: event branch %s does not match deployed branch %s\n",
			branch, deployment.GetBranch())
		return
	}

	// If branches match, deploy
	fmt.Fprintf(out, "Accepting event: event branch %s matches deployed branch %s\n",
		branch, deployment.GetBranch())
	deploy, err := deployment.Deploy(cli, os.Stdout, project.DeployOptions{})
	if err != nil {
		fmt.Fprintln(out, "Build failed: "+err.Error())
	}

	err = deploy()
	if err != nil {
		fmt.Fprintln(out, "Deploy failed: "+err.Error())
	}
}
