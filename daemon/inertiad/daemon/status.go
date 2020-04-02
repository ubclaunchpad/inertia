package daemon

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/blang/semver"
	"github.com/go-chi/render"

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
	status, err := s.deployment.GetStatus(s.docker)
	status.InertiaVersion = s.version

	// badge generator for https://shields.io/endpoint
	if r.URL.Query().Get("badge") == "true" {
		badge := shieldsIOData{1, "inertia", "blue", "deployed", "green", false}
		if err != nil {
			badge.Message = "errored"
			badge.Color = "red"
			badge.IsError = true
			render.JSON(w, r, badge)
			return
		} else if status.BuildContainerActive {
			// build in progress
			badge.Message = "deploying"
			badge.Color = "yellow"
			render.JSON(w, r, badge)
			return
		} else if status.CommitHash == "" {
			// no project
			badge.Message = "no project"
			badge.Color = "lightgrey"
			badge.IsError = true
			render.JSON(w, r, badge)
			return
		} else if len(status.Containers) == 0 {
			// deployment offline
			badge.Message = "project offline"
			badge.Color = "grey"
			badge.IsError = true
			render.JSON(w, r, badge)
			return
		}

		// deployed
		render.JSON(w, r, badge)
		return
	}

	// check for new version
	current, _ := semver.Parse(strings.TrimPrefix(s.version, "v"))
	latest, tagCheckErr := containers.GetLatestImageTag(r.Context(), "ubclaunchpad/inertia", &current)
	if tagCheckErr == nil {
		verStr := fmt.Sprintf("v%s", latest.String())
		status.NewVersionAvailable = &verStr
	}

	// standard responses
	if status.CommitHash == "" {
		status.Containers = make([]string, 0)
		render.Render(w, r, res.MsgOK("status retrieved",
			"status", status))
		return
	}
	if err != nil {
		render.Render(w, r, res.ErrInternalServer("failed to get status", err))
		return
	}
	render.Render(w, r, res.MsgOK("status retrieved",
		"status", status))
}
