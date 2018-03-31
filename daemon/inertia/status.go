package main

import (
	"fmt"
	"net/http"

	docker "github.com/docker/docker/client"
)

const (
	msgBuildInProgress    = "It appears that your build is still in progress."
	msgNoContainersActive = "No containers are active."
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

	branchStatus := " - Branch:  " + status.Branch + "\n"
	commitStatus := " - Commit:  " + status.CommitHash + "\n"
	commitMessage := " - Message: " + status.CommitMessage + "\n"
	statusString := inertiaStatus + branchStatus + commitStatus + commitMessage

	// If build container is active, that means that a build
	// attempt was made but only the daemon and docker-compose
	// are active, indicating a build failure or build-in-progress
	if len(status.Containers) == 0 {
		if status.BuildContainerActive {
			errorString := statusString + msgBuildInProgress
			http.Error(w, errorString, http.StatusOK)
		} else {
			errorString := statusString + msgNoContainersActive
			http.Error(w, errorString, http.StatusOK)
		}
		return
	}

	activeContainers := "Active containers:\n"
	for _, container := range status.Containers {
		activeContainers += " - " + container + "\n"
	}
	statusString += activeContainers

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, statusString)
}
