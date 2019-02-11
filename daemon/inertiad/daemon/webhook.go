package daemon

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-chi/render"
	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/webhook"
)

// webhookHandler receives and parses Git-based webhooks
// Supported vendors: Github, Gitlab, Bitbucket
// Supported events: push
func (s *Server) webhookHandler(w http.ResponseWriter, r *http.Request) {
	// read
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := "unable to read payload: " + err.Error()
		println(msg)
		render.Render(w, r, res.ErrBadRequest(msg))
		return
	}

	// check type
	host, event := webhook.Type(r.Header)

	// ensure validity
	if s.state.WebhookSecret == "" {
		println("warning: no webhook secret is set up yet! set one in inertia.toml and run inertia [remote] up")
	}
	if err := webhook.Verify(host, s.state.WebhookSecret, r.Header, body); err != nil {
		msg := "unable to verify payload: " + err.Error()
		println(msg)
		render.Render(w, r, res.ErrBadRequest(msg))
		return
	}

	// retrieve payload
	payload, err := webhook.Parse(host, event, r.Header, body)
	if err != nil {
		msg := "unable to parse payload: " + err.Error()
		println(msg)
		render.Render(w, r, res.ErrBadRequest(msg))
		return
	}

	// process event
	switch event := payload.GetEventType(); event {
	case webhook.PingEvent:
		fmt.Fprint(w, api.MsgDaemonOK)
		println("ping webhook received")
		return
	case webhook.PushEvent:
		fmt.Fprint(w, api.MsgDaemonOK)
		processPushEvent(s, payload)
	// case webhook.PullEvent:
	//	fmt.Fprint(w, common.MsgDaemonOK)
	// 	processPullRequestEvent(payload)
	default:
		render.Render(w, r, res.ErrBadRequest("unrecognized event type"))
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
func processPushEvent(s *Server, p webhook.Payload) {
	fmt.Printf("Received %s push event: %s (%s)\n",
		p.GetSource(), p.GetRepoName(), p.GetRef())

	// Ignore event if repository not set up yet, otherwise
	// let deploy() handle the update.
	if status, _ := s.deployment.GetStatus(s.docker); status.CommitHash == "" {
		fmt.Println(msgNoDeployment)
		return
	}

	// Check for matching remotes
	if err := s.deployment.CompareRemotes(p.GetSSHURL()); err != nil {
		fmt.Println(err.Error())
		return
	}

	// Check for matching branch
	var branch = common.GetBranchFromRef(p.GetRef())
	if s.deployment.GetBranch() != branch {
		fmt.Printf("Ignoring event: event branch %s does not match deployed branch %s\n",
			branch, s.deployment.GetBranch())
		return
	}

	// If branches match, deploy
	fmt.Printf("Accepting event: event branch %s matches deployed branch %s\n",
		branch, s.deployment.GetBranch())
	deploy, err := s.deployment.Deploy(s.docker, os.Stdout, project.DeployOptions{})
	if err != nil {
		fmt.Println("Build failed: " + err.Error())
		return
	}

	if err = deploy(); err != nil {
		fmt.Println("Deploy failed: " + err.Error())
	}
}
