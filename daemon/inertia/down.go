package main

import (
	"fmt"
	"net/http"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/daemon/inertia/project"
)

// downHandler tries to take the deployment offline
func downHandler(w http.ResponseWriter, r *http.Request) {
	println("DOWN request received")

	logger := newLogger(false, w)
	defer logger.Close()

	cli, err := docker.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	// Error if no project containers are active, but try to kill
	// everything anyway in case the docker-compose image is still
	// active
	_, err = project.GetActiveContainers(cli)
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		err = project.KillActiveContainers(cli, logger.GetWriter())
		if err != nil {
			println(err)
		}
		return
	}

	err = project.KillActiveContainers(cli, logger.GetWriter())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Project shut down.")
}
