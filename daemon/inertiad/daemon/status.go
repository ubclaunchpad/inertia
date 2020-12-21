package daemon

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/blang/semver"
	"github.com/go-chi/render"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

type shieldsIOData struct {
	SchemaVersion int `json:"schemaVersion"` // always 1

	Label      string `json:"label"`      // left text
	LabelColor string `json:"labelColor"` // left color

	Message string `json:"message"` // right text
	Color   string `json:"color"`   // right color

	IsError bool `json:"isError"`
}

// statusHandler returns a formatted string about the status of the
// deployment and lists currently active project containers
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	status, statusErr := s.deployment.GetStatus(s.docker)
	extendedStatus := &api.DeploymentStatusWithVersions{
		DeploymentStatus: status,
		InertiaVersion:   s.version,
	}

	// badge generator for https://shields.io/endpoint
	if r.URL.Query().Get("badge") == "true" {
		render.JSON(w, r, generateBadge(statusErr, status))
		return
	}

	// set up message
	okMsg := "status retrieved"

	// check for new version
	current, _ := semver.Parse(strings.TrimPrefix(s.version, "v"))
	latest, tagCheckErr := containers.GetLatestImageTag(r.Context(), containers.Image{
		Registry: "ghcr.io",
		// check inertia repository instead of inertiad, since ghcr.io has no REST API yet
		Repository: "ubclaunchpad/inertia",
	}, &current)
	if tagCheckErr != nil {
		okMsg = fmt.Sprintf("%s, but failed to fetch Inertia updates: %s", okMsg, tagCheckErr)
	} else if tagCheckErr == nil && current.LT(*latest) {
		verStr := fmt.Sprintf("v%s", latest.String())
		extendedStatus.NewVersionAvailable = &verStr
	}

	// standard responses
	if status.CommitHash == "" {
		extendedStatus.Containers = make([]string, 0)
		render.Render(w, r, res.MsgOK(okMsg, "status", extendedStatus))
		return
	}
	if statusErr != nil {
		render.Render(w, r, res.ErrInternalServer("failed to get status", statusErr))
		return
	}
	render.Render(w, r, res.MsgOK(okMsg, "status", extendedStatus))
}

func generateBadge(statusErr error, status api.DeploymentStatus) shieldsIOData {
	badge := shieldsIOData{1, "inertia", "blue", "deployed", "green", false}
	if statusErr != nil {
		badge.Message = "errored"
		badge.Color = "red"
		badge.IsError = true
	} else if status.BuildContainerActive {
		// build in progress
		badge.Message = "deploying"
		badge.Color = "yellow"
	} else if status.CommitHash == "" {
		// no project
		badge.Message = "no project"
		badge.Color = "lightgrey"
		badge.IsError = true
	} else if len(status.Containers) == 0 {
		// deployment offline
		badge.Message = "project offline"
		badge.Color = "grey"
		badge.IsError = true
	}
	return badge
}
