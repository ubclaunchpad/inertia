package main

import (
	"encoding/json"
	"net/http"

	docker "github.com/docker/docker/client"
)

// statusHandler returns a formatted string about the status of the
// deployment and lists currently active project containers
func statusHandler(w http.ResponseWriter, r *http.Request) {
	inertiaStatus := "inertia daemon " + daemonVersion + "\n"
	if deployment == nil {
		http.Error(
			w, inertiaStatus+msgNoDeployment,
			http.StatusNotFound,
		)
		return
	}

	cli, err := docker.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	status, err := deployment.GetStatus(cli)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
