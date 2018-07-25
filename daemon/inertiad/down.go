package main

import (
	"net/http"
	"os"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
)

// downHandler tries to take the deployment offline
func downHandler(w http.ResponseWriter, r *http.Request) {
	if deployment == nil {
		http.Error(w, msgNoDeployment, http.StatusPreconditionFailed)
		return
	}

	logger := log.NewLogger(log.LoggerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})
	defer logger.Close()

	cli, err := containers.NewDockerClient()
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusPreconditionFailed)
		return
	}
	defer cli.Close()

	err = deployment.Down(cli, logger)
	if err == containers.ErrNoContainers {
		logger.WriteErr(err.Error(), http.StatusPreconditionFailed)
		return
	} else if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WriteSuccess("Project shut down.", http.StatusOK)
}
