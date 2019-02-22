package daemon

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
)

// upHandler tries to bring the deployment online
func (s *Server) upHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var upReq api.UpRequest
	if err = json.Unmarshal(body, &upReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var gitOpts = upReq.GitOptions

	// apply configuration updates
	s.state.WebhookSecret = upReq.WebHookSecret
	s.deployment.SetConfig(project.DeploymentConfig{
		ProjectName:   upReq.Project,
		BuildType:     upReq.BuildType,
		BuildFilePath: upReq.BuildFilePath,
		RemoteURL:     gitOpts.RemoteURL,
		Branch:        gitOpts.Branch,
	})

	// Configure logger
	logger := log.NewLogger(log.LoggerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
		HTTPStream: upReq.Stream,
	})
	defer logger.Close()

	// Check for existing git repository, clone if no git repository exists.
	var skipUpdate = false
	if status, _ := s.deployment.GetStatus(s.docker); status.CommitHash == "" {
		logger.Println("No deployment detected")
		if err = s.deployment.Initialize(
			project.DeploymentConfig{
				ProjectName:   upReq.Project,
				BuildType:     upReq.BuildType,
				BuildFilePath: upReq.BuildFilePath,
				RemoteURL:     gitOpts.RemoteURL,
				Branch:        gitOpts.Branch,
				PemFilePath:   crypto.DaemonGithubKeyLocation,
			},
			logger,
		); err != nil {
			logger.WriteErr(err.Error(), http.StatusPreconditionFailed)
			return
		}

		// Project was just pulled! No need to update again.
		skipUpdate = true
	}

	// Check for matching remotes
	if err = s.deployment.CompareRemotes(gitOpts.RemoteURL); err != nil {
		logger.WriteErr(err.Error(), http.StatusPreconditionFailed)
		return
	}

	// Change deployment parameters if necessary
	s.deployment.SetConfig(project.DeploymentConfig{
		ProjectName: upReq.Project,
		Branch:      gitOpts.Branch,
	})

	// Deploy project
	deploy, err := s.deployment.Deploy(s.docker, logger, project.DeployOptions{
		SkipUpdate: skipUpdate,
	})
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	if err = deploy(); err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	s.deployment.UpdateContainerHistory(s.docker)

	logger.WriteSuccess("Project startup initiated!", http.StatusCreated)
}
