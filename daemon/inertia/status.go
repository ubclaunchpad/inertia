package main

import (
	"fmt"
	"net/http"

	docker "github.com/docker/docker/client"
	git "gopkg.in/src-d/go-git.v4"
)

// statusHandler lists currently active project containers
func statusHandler(w http.ResponseWriter, r *http.Request) {
	println("STATUS request received")

	inertiaStatus := "inertia daemon " + daemonVersion + "\n"

	// Get status of repository
	repo, err := git.PlainOpen(projectDirectory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		return
	}
	head, err := repo.Head()
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		return
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return
	}
	branchStatus := " - Branch:  " + head.Name().Short() + "\n"
	commitStatus := " - Commit:  " + head.Hash().String() + "\n"
	commitMessage := " - Message: " + commit.Message + "\n"
	status := inertiaStatus + branchStatus + commitStatus + commitMessage

	// Get containers
	cli, err := docker.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	containers, err := getActiveContainers(cli)
	if err != nil {
		if err.Error() == noContainersResp {
			// This is different from having 2 containers active -
			// noContainersResp means that no attempt to build the project
			// was made or the project was cleanly shut down.
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, status+noContainersResp)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If there are only 2 containers active, that means that a build
	// attempt was made but only the daemon and the docker-compose containers
	// are active, indicating a build failure.
	if len(containers) == 2 {
		errorString := status + "It appears that an attempt to start your project was made but the build failed."
		http.Error(w, errorString, http.StatusNotFound)
		return
	}

	ignore := map[string]bool{
		"/inertia-daemon": true,
		"/docker-compose": true,
	}
	// Only list project containers
	activeContainers := "Active containers:"
	for _, container := range containers {
		if !ignore[container.Names[0]] {
			activeContainers += "\n" + container.Image + " (" + container.Names[0] + ")"
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, status+activeContainers)
}
