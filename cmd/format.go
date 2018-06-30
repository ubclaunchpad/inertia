package cmd

import (
	"fmt"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/common"
)

const (
	msgBuildInProgress    = "It appears that your build is still in progress."
	msgNoContainersActive = "No containers are active."
	msgNoDeployment       = "No deployment found - try running 'inertia [VPS] up'"
)

func formatStatus(s *common.DeploymentStatus) string {
	inertiaStatus := "inertia daemon " + s.InertiaVersion + "\n"
	branchStatus := " - Branch:     " + s.Branch + "\n"
	commitStatus := " - Commit:     " + s.CommitHash + "\n"
	commitMessage := " - Message:    " + s.CommitMessage + "\n"
	buildTypeStatus := " - Build Type: " + s.BuildType + "\n"
	statusString := inertiaStatus + branchStatus + commitStatus + commitMessage + buildTypeStatus

	// If no branch/commit, then it's likely the deployment has not
	// been instantiated on the remote yet
	if s.Branch == "" && s.CommitHash == "" && s.CommitMessage == "" {
		return statusString + msgNoDeployment
	}

	// If build container is active, that means that a build
	// attempt was made but only the daemon and docker-compose
	// are active, indicating a build failure or build-in-progress
	if len(s.Containers) == 0 {
		errorString := statusString + msgNoContainersActive
		if s.BuildContainerActive {
			errorString = statusString + msgBuildInProgress
		}
		return errorString
	}

	activeContainers := "Active containers:\n"
	for _, container := range s.Containers {
		activeContainers += " - " + container + "\n"
	}
	statusString += activeContainers
	return statusString
}

func formatRemoteDetails(remote *cfg.RemoteVPS) string {
	remoteString := fmt.Sprintf("Remote %s: \n", remote.Name)
	remoteString += fmt.Sprintf(" - Deployed Branch:   %s\n", remote.Branch)
	remoteString += fmt.Sprintf(" - IP Address:        %s\n", remote.IP)
	remoteString += fmt.Sprintf(" - VPS User:          %s\n", remote.User)
	remoteString += fmt.Sprintf(" - PEM File Location: %s\n", remote.PEM)
	remoteString += fmt.Sprintf("Run 'inertia %s status' for more details.\n", remote.Name)
	return remoteString
}
