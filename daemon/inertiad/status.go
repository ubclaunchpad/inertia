package main

import (
	"encoding/json"
	"net/http"

	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
)

// statusHandler returns a formatted string about the status of the
// deployment and lists currently active project containers
func statusHandler(w http.ResponseWriter, r *http.Request) {
	if deployment == nil {
		status := &common.DeploymentStatus{
			InertiaVersion: Version,
			Containers:     make([]string, 0),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(status)
		return
	}

	cli, err := containers.NewDockerClient()
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
	status.InertiaVersion = Version

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
