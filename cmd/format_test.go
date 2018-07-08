package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/common"
)

func TestFormatStatus(t *testing.T) {
	output := formatStatus(&common.DeploymentStatus{
		InertiaVersion:       "9000",
		Branch:               "call",
		CommitHash:           "me",
		CommitMessage:        "maybe",
		BuildContainerActive: true,
		Containers:           []string{"wow"},
	})
	assert.Contains(t, output, "inertia daemon 9000")
	assert.Contains(t, output, "Active containers")
}

func TestFormatStatusBuildActive(t *testing.T) {
	output := formatStatus(&common.DeploymentStatus{
		InertiaVersion:       "9000",
		Branch:               "call",
		CommitHash:           "me",
		CommitMessage:        "maybe",
		BuildContainerActive: true,
		Containers:           make([]string, 0),
	})
	assert.Contains(t, output, "inertia daemon 9000")
	assert.Contains(t, output, msgBuildInProgress)
}

func TestFormatStatusNoDeployment(t *testing.T) {
	output := formatStatus(&common.DeploymentStatus{
		InertiaVersion:       "9000",
		Branch:               "",
		CommitHash:           "",
		CommitMessage:        "",
		BuildContainerActive: false,
		Containers:           make([]string, 0),
	})
	assert.Contains(t, output, "inertia daemon 9000")
	assert.Contains(t, output, msgNoDeployment)
}

func TestFormatRemoteDetails(t *testing.T) {
	client := &cfg.RemoteVPS{
		Name:   "bob",
		Branch: "great",
		User:   "tree",
		PEM:    "/wow/amaze",
	}
	output := formatRemoteDetails(client)
	assert.Contains(t, output, "bob")
	assert.Contains(t, output, "great")
	assert.Contains(t, output, "tree")
	assert.Contains(t, output, "/wow/amaze")
}
