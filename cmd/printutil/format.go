package printutil

import (
	"fmt"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/cfg"
)

const (
	msgBuildInProgress    = "It appears that your build is still in progress."
	msgNoContainersActive = "No containers are active."
	msgNoDeployment       = "No deployment found - try running 'inertia [remote] up'"
)

// FormatStatus prints the given deployment status
func FormatStatus(s *api.DeploymentStatus) string {
	var (
		inertiaStatus   = "inertia daemon " + s.InertiaVersion + "\n"
		branchStatus    = " - Branch:     " + s.Branch + "\n"
		commitStatus    = " - Commit:     " + s.CommitHash + "\n"
		commitMessage   = " - Message:    " + s.CommitMessage + "\n"
		buildTypeStatus = " - Build Type: " + s.BuildType + "\n"
	)

	// If no branch/commit, then it's likely the deployment has not
	// been instantiated on the remote yet
	var statusString = inertiaStatus + branchStatus + commitStatus + commitMessage + buildTypeStatus
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

// FormatRemoteDetails prints the given remote configuration
func FormatRemoteDetails(name string, remote *cfg.Remote) string {
	remoteString := fmt.Sprintf("Remote %s: \n", name)
	remoteString += fmt.Sprintf(" - IP Address:        %s\n", remote.IP)
	if remote.SSH != nil {
		remoteString += fmt.Sprintf(" - VPS User:          %s\n", remote.SSH.User)
		remoteString += fmt.Sprintf(" - PEM File Location: %s\n", remote.SSH.PEM)
	} else {
		remoteString += " - VPS User:\n - PEM File Location:\n"
	}
	remoteString += fmt.Sprintf("Run 'inertia %s status' for more details.\n", name)
	return remoteString
}
