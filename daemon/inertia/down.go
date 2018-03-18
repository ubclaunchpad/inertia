package main

import (
	"fmt"
	"net/http"

	docker "github.com/docker/docker/client"
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
	_, err = getActiveContainers(cli)
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		err = killActiveContainers(cli, logger.GetWriter())
		if err != nil {
			println(err)
		}
		return
	}

	err = killActiveContainers(cli, logger.GetWriter())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Project shut down.")
}
