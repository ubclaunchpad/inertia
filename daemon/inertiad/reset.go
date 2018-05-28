package main

import (
	"net/http"
	"os"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
)

// resetHandler shuts down and wipes the project directory
func resetHandler(w http.ResponseWriter, r *http.Request) {
	if deployment == nil {
		http.Error(w, msgNoDeployment, http.StatusPreconditionFailed)
		return
	}

	logger := log.NewLogger(os.Stdout, nil, w)
	defer logger.Close()

	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	// Goodbye deployment
	err = deployment.Destroy(cli, logger)
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}
	deployment = nil

	logger.WriteSuccess("Project removed from remote.", http.StatusOK)
}
