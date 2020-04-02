package out

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
func FormatStatus(remoteName string, s *api.DeploymentStatus) string {
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

	// report new version if one is available
	if s.NewVersionAvailable != nil && *s.NewVersionAvailable != "" {
		statusString += C("\n:rocket: Inertia version %s is now available!\n", CY).
			With(*s.NewVersionAvailable).String()
		statusString += fmt.Sprintf("Go to https://github.com/ubclaunchpad/inertia/releases/tag/%s for more details.\n",
			*s.NewVersionAvailable)
		statusString += fmt.Sprintf("Run 'inertia %s upgrade --help' for tips on upgrading.\n",
			remoteName)
	}

	return statusString
}

// FormatRemoteDetails prints the given remote configuration
func FormatRemoteDetails(remote cfg.Remote) string {
	var remoteString string
	remoteString += fmt.Sprintf(":globe_with_meridians: IP:                   %s\n", remote.IP)
	remoteString += fmt.Sprintf(":airplane: Version:              %s\n", remote.Version)
	if remote.Daemon != nil {
		remoteString += fmt.Sprintf(":passport_control: Daemon.Port:          %s\n", remote.Daemon.Port)
		remoteString += fmt.Sprintf(":lock: Daemon.Authenticated: %v\n", remote.Daemon.Token != "")
		remoteString += fmt.Sprintf(":mag: Daemon.VerifySSL:     %v\n", remote.Daemon.VerifySSL)
	}
	if remote.SSH != nil {
		remoteString += fmt.Sprintf(":ghost: SSH.User:             %s\n", remote.SSH.User)
		remoteString += fmt.Sprintf(":key: SSH.IdentityFile:     %s\n", remote.SSH.IdentityFile)
		remoteString += fmt.Sprintf(":customs: SSH.Port:             %s\n", remote.SSH.SSHPort)
	}
	if remote.Profiles != nil {
		remoteString += fmt.Sprintf(":open_file_folder: Profiles:             %v", remote.Profiles)
	}
	return remoteString
}
