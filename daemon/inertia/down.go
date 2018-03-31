package main

import (
	"net/http"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/daemon/inertia/project"
)

// downHandler tries to take the deployment offline
func downHandler(w http.ResponseWriter, r *http.Request) {
	if deployment == nil {
		http.Error(w, msgNoDeployment, http.StatusPreconditionFailed)
		return
	}

	logger := newLogger(false, w)
	defer logger.Close()

	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.Err(err.Error(), http.StatusPreconditionFailed)
		return
	}
	defer cli.Close()

	err = deployment.Down(cli, logger.GetWriter())
	if err == project.ErrNoContainers {
		logger.Err(err.Error(), http.StatusPreconditionFailed)
		return
	} else if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Success("Project shut down.", http.StatusOK)
}
