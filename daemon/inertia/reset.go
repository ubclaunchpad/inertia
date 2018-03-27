package main

import (
	"net/http"

	docker "github.com/docker/docker/client"
)

// resetHandler shuts down and wipes the project directory
func resetHandler(w http.ResponseWriter, r *http.Request) {
	if deployment == nil {
		http.Error(w, noDeploymentMsg, http.StatusPreconditionFailed)
		return
	}

	logger := newLogger(false, w)
	defer logger.Close()

	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	// Goodbye deployment
	err = deployment.Destroy(cli)
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}
	deployment = nil

	logger.Success("Project removed from remote.", http.StatusOK)
}
