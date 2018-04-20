package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertia/auth"
	"github.com/ubclaunchpad/inertia/daemon/inertia/project"
)

// upHandler tries to bring the deployment online
func upHandler(w http.ResponseWriter, r *http.Request) {
	// Get github URL from up request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var upReq common.DaemonRequest
	err = json.Unmarshal(body, &upReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger := newLogger(upReq.Stream, w)
	gitOpts := upReq.GitOptions
	defer logger.Close()

	webhookSecret = upReq.Secret

	// Check for existing git repository, clone if no git repository exists.
	skipUpdate := false
	if deployment == nil {
		logger.Println("No deployment detected")
		common.RemoveContents(project.Directory)
		d, err := project.NewDeployment(project.DeploymentConfig{
			ProjectName: upReq.Project,
			BuildType:   upReq.BuildType,
			RemoteURL:   gitOpts.RemoteURL,
			Branch:      gitOpts.Branch,
			PemFilePath: auth.DaemonGithubKeyLocation,
		}, logger.GetWriter())
		if err != nil {
			logger.Err(err.Error(), http.StatusPreconditionFailed)
			return
		}
		deployment = d

		// Project was just pulled! No need to update again.
		skipUpdate = true
	}

	// Check for matching remotes
	err = deployment.CompareRemotes(gitOpts.RemoteURL)
	if err != nil {
		logger.Err(err.Error(), http.StatusPreconditionFailed)
		return
	}

	// Change deployment parameters if necessary
	deployment.SetConfig(project.DeploymentConfig{
		ProjectName: upReq.Project,
		Branch:      gitOpts.Branch,
	})

	// Deploy project
	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	err = deployment.Deploy(cli, logger.GetWriter(), project.DeployOptions{
		SkipUpdate: skipUpdate,
	})
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Success("Project startup initiated!", http.StatusCreated)
}
