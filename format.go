package main

import (
	"github.com/ubclaunchpad/inertia/common"
)

const (
	msgBuildInProgress    = "It appears that your build is still in progress."
	msgNoContainersActive = "No containers are active."
)

func formatStatus(status *common.DeploymentStatus) string {
	inertiaStatus := "inertia daemon " + status.InertiaVersion + "\n"
	branchStatus := " - Branch:     " + status.Branch + "\n"
	commitStatus := " - Commit:     " + status.CommitHash + "\n"
	commitMessage := " - Message:    " + status.CommitMessage + "\n"
	buildTypeStatus := " - Build Type: " + status.BuildType + "\n"
	statusString := inertiaStatus + branchStatus + commitStatus + commitMessage + buildTypeStatus

	// If build container is active, that means that a build
	// attempt was made but only the daemon and docker-compose
	// are active, indicating a build failure or build-in-progress
	if len(status.Containers) == 0 {
		errorString := statusString + msgNoContainersActive
		if status.BuildContainerActive {
			errorString = statusString + msgBuildInProgress
		}
		return errorString
	}

	activeContainers := "Active containers:\n"
	for _, container := range status.Containers {
		activeContainers += " - " + container + "\n"
	}
	statusString += activeContainers
	return statusString
}
