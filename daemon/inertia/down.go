package main

import (
	"net/http"
	"os"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/daemon/inertia/log"
	"github.com/ubclaunchpad/inertia/daemon/inertia/project"
)

// downHandler tries to take the deployment offline
func downHandler(w http.ResponseWriter, r *http.Request) {
	if deployment == nil {
		http.Error(w, msgNoDeployment, http.StatusPreconditionFailed)
		return
	}

	logger := log.NewLogger(os.Stdout, nil, w)
	defer logger.Close()

	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusPreconditionFailed)
		return
	}
	defer cli.Close()

	err = deployment.Down(cli, logger)
	if err == project.ErrNoContainers {
		logger.WriteErr(err.Error(), http.StatusPreconditionFailed)
		return
	} else if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WriteSuccess("Project shut down.", http.StatusOK)
}
