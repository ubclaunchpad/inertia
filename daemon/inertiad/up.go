package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/ubclaunchpad/inertia/common"
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

	var (
		upReq   common.UpRequest
		gitOpts common.GitOptions
	)

	if err := json.Unmarshal(body, &upReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	webhookSecret = upReq.WebHookSecret
	if upReq.GitOptions != nil {
		gitOpts = *upReq.GitOptions
	}

	// Configure logger
	logger := log.NewLogger(log.LoggerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
		HTTPStream: upReq.Stream,
	})
	defer logger.Close()

	// Parse project configuration
	projectConfig, err := common.ReadProjectConfig(path.Join(conf.ProjectDirectory, "inertia.toml"))
	if err != nil {
		logger.WriteErr("Failed to read project configuration", http.StatusPreconditionFailed)
		return
	}

	// Set up Docker
	cli, err := containers.NewDockerClient()
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	// Check for existing git repository, clone if no git repository exists.
	skipUpdate := false
	if status, _ := deployment.GetStatus(cli); status.CommitHash == "" {
		logger.Println("No deployment detected")
		err = deployment.Initialize(
			project.DeploymentConfig{
				ProjectName:   common.Dereference(projectConfig.Project),
				BuildType:     common.Dereference(projectConfig.Build.Type),
				BuildFilePath: common.Dereference(projectConfig.Build.ConfigPath),
				RemoteURL:     common.Dereference(projectConfig.Repository.RemoteURL),
				Branch:        gitOpts.Branch,
				PemFilePath:   crypto.DaemonGithubKeyLocation,
			},
			logger)
		if err != nil {
			logger.WriteErr(err.Error(), http.StatusPreconditionFailed)
			return
		}

		// Project was just pulled! No need to update again.
		skipUpdate = true
	}

	// Check for matching remotes
	if err := deployment.CompareRemotes(common.Dereference(projectConfig.Repository.RemoteURL)); err != nil {
		logger.WriteErr(err.Error(), http.StatusPreconditionFailed)
		return
	}

	// Change deployment parameters if necessary
	deployment.SetConfig(project.DeploymentConfig{
		Branch: gitOpts.Branch,
	})

	// Build project
	deploy, err := deployment.Deploy(cli, logger, project.DeployOptions{
		SkipUpdate: skipUpdate,
	})
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	// Deploy project
	if err := deploy(); err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WriteSuccess("Project startup initiated!", http.StatusCreated)
}
