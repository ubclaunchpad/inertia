package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertia/auth"
	"github.com/ubclaunchpad/inertia/daemon/inertia/log"
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
	gitOpts := upReq.GitOptions

	// Upgrade to websocket connection if required
	var logger *log.DaemonLogger
	if upReq.Stream {
		socket, err := socketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger = log.NewLogger(os.Stdout, socket, w)
	} else {
		logger = log.NewLogger(os.Stdout, nil, w)
	}
	defer logger.Close()

	webhookSecret = upReq.Secret

	// Check for existing git repository, clone if no git repository exists.
	skipUpdate := false
	if deployment == nil {
		logger.Println("No deployment detected")
		d, err := project.NewDeployment(project.DeploymentConfig{
			ProjectName: upReq.Project,
			BuildType:   upReq.BuildType,
			RemoteURL:   gitOpts.RemoteURL,
			Branch:      gitOpts.Branch,
			PemFilePath: auth.DaemonGithubKeyLocation,
		}, logger)
		if err != nil {
			logger.WriteErr(err.Error(), http.StatusPreconditionFailed)
			return
		}
		deployment = d

		// Project was just pulled! No need to update again.
		skipUpdate = true
	}

	// Check for matching remotes
	err = deployment.CompareRemotes(gitOpts.RemoteURL)
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusPreconditionFailed)
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
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	err = deployment.Deploy(cli, logger, project.DeployOptions{
		SkipUpdate: skipUpdate,
	})
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WriteSuccess("Project startup initiated!", http.StatusCreated)
}
