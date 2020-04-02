package out

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/cfg"
)

func TestFormatStatus(t *testing.T) {
	out := FormatStatus("robert", &api.DeploymentStatus{
		InertiaVersion:       "9000",
		Branch:               "call",
		CommitHash:           "me",
		CommitMessage:        "maybe",
		BuildContainerActive: true,
		Containers:           []string{"wow"},
	})
	assert.Contains(t, out, "inertia daemon 9000")
	assert.Contains(t, out, "Active containers")

	t.Run("with build active", func(t *testing.T) {
		out := FormatStatus("robert", &api.DeploymentStatus{
			InertiaVersion:       "9000",
			Branch:               "call",
			CommitHash:           "me",
			CommitMessage:        "maybe",
			BuildContainerActive: true,
			Containers:           make([]string, 0),
		})
		assert.Contains(t, out, "inertia daemon 9000")
		assert.Contains(t, out, msgBuildInProgress)
	})

	t.Run("with new version available", func(t *testing.T) {
		version := "v0.6.0"
		out := FormatStatus("robert", &api.DeploymentStatus{
			InertiaVersion:       "9000",
			Branch:               "call",
			CommitHash:           "me",
			CommitMessage:        "maybe",
			BuildContainerActive: true,
			Containers:           []string{"wow"},
			NewVersionAvailable:  &version,
		})
		assert.Contains(t, out, "inertia daemon 9000")
		assert.Contains(t, out, version)
		assert.Contains(t, out, "robert")
	})
}

func TestFormatStatusNoDeployment(t *testing.T) {
	out := FormatStatus("robert", &api.DeploymentStatus{
		InertiaVersion:       "9000",
		Branch:               "",
		CommitHash:           "",
		CommitMessage:        "",
		BuildContainerActive: false,
		Containers:           make([]string, 0),
	})
	assert.Contains(t, out, "inertia daemon 9000")
	assert.Contains(t, out, msgNoDeployment)
}

func TestFormatRemoteDetails(t *testing.T) {
	var out = FormatRemoteDetails(cfg.Remote{
		Name: "bob",
		SSH: &cfg.SSH{
			User:         "tree",
			IdentityFile: "/wow/amaze",
		},
	})
	assert.Contains(t, out, "tree")
	assert.Contains(t, out, "/wow/amaze")
	out = FormatRemoteDetails(cfg.Remote{Name: "bob", IP: "0.0.0.0"})
	assert.Contains(t, out, "0.0.0.0")
}
