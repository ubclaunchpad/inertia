package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/build"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
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
	var upReq common.UpRequest
	err = json.Unmarshal(body, &upReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gitOpts := upReq.GitOptions

	// Configure logger
	logger := log.NewLogger(log.LoggerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
		HTTPStream: upReq.Stream,
	})
	defer logger.Close()

	webhookSecret = upReq.WebHookSecret

	// Check for existing git repository, clone if no git repository exists.
	skipUpdate := false
	if deployment == nil {
		logger.Println("No deployment detected")
		d, err := project.NewDeployment(
			build.NewBuilder(*conf, containers.StopActiveContainers),
			project.DeploymentConfig{
				ProjectDirectory: conf.ProjectDirectory,
				ProjectName:      upReq.Project,
				BuildType:        upReq.BuildType,
				BuildFilePath:    upReq.BuildFilePath,
				RemoteURL:        gitOpts.RemoteURL,
				Branch:           gitOpts.Branch,
				PemFilePath:      crypto.DaemonGithubKeyLocation,
				DatabasePath:     path.Join(conf.DataDirectory, "project.db"),
			},
			logger,
		)
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
