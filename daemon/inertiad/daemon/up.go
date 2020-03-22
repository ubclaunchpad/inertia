package daemon

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-chi/render"
	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

// upHandler tries to bring the deployment online
func (s *Server) upHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.Render(w, r, res.ErrBadRequest(err.Error()))
		return
	}
	defer r.Body.Close()
	var upReq api.UpRequest
	if err = json.Unmarshal(body, &upReq); err != nil {
		render.Render(w, r, res.ErrBadRequest(err.Error()))
		return
	}
	var gitOpts = upReq.GitOptions

	// apply configuration updates
	if upReq.WebHookSecret != "" {
		s.state.WebhookSecret = upReq.WebHookSecret
	}
	conf := project.DeploymentConfig{
		ProjectName:            upReq.Project,
		BuildType:              upReq.BuildType,
		BuildFilePath:          upReq.BuildFilePath,
		RemoteURL:              gitOpts.RemoteURL,
		Branch:                 gitOpts.Branch,
		PemFilePath:            crypto.DaemonGithubKeyLocation,
		IntermediaryContainers: upReq.IntermediaryContainers,
		SlackNotificationURL:   upReq.SlackNotificationURL,
	}
	s.deployment.SetConfig(conf)

	// Configure streamer
	var stream = log.NewStreamer(log.StreamerOptions{
		Request:    r,
		Stdout:     os.Stdout,
		HTTPWriter: w,
		HTTPStream: upReq.Stream,
	})
	defer stream.Close()

	// Check for existing git repository, clone if no git repository exists.
	var skipUpdate = false
	if status, _ := s.deployment.GetStatus(s.docker); status.CommitHash == "" {
		stream.Println("No deployment detected")
		if err = s.deployment.Initialize(conf, stream); err != nil {
			stream.Error(res.Err(err.Error(), http.StatusPreconditionFailed))
			return
		}

		// Project was just pulled! No need to update again.
		skipUpdate = true
	}

	// Check for matching remotes
	if err = s.deployment.CompareRemotes(gitOpts.RemoteURL); err != nil {
		stream.Error(res.Err(err.Error(), http.StatusPreconditionFailed))
		return
	}

	// Change deployment parameters if necessary
	s.deployment.SetConfig(project.DeploymentConfig{
		ProjectName: upReq.Project,
		Branch:      gitOpts.Branch,
	})

	// Deploy project
	deploy, err := s.deployment.Deploy(s.docker, stream, project.DeployOptions{
		SkipUpdate: skipUpdate,
	})
	if err != nil {
		stream.Error(res.ErrInternalServer("failed to build project", err))
		return
	}

	if err = deploy(); err != nil {
		stream.Error(res.ErrInternalServer("failed to deploy project", err))
		return
	}

	// Update container management history following a successful build and deployment
	if err = s.deployment.UpdateContainerHistory(s.docker); err != nil {
		stream.Println("warning: failed to update container history:", err)
	}

	stream.Success(res.Msg("Project startup initiated!", http.StatusCreated))
}
