package main

import (
	"net/http"
	"os"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
)

// pruneHandler cleans up Docker assets
func pruneHandler(w http.ResponseWriter, r *http.Request) {
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
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	err = deployment.Prune(cli, logger)
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}
	logger.WriteSuccess("Docker assets have been pruned.", http.StatusOK)
}
