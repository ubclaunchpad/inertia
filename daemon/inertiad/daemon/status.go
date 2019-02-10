package daemon

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

// statusHandler returns a formatted string about the status of the
// deployment and lists currently active project containers
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	status, err := s.deployment.GetStatus(s.docker)
	status.InertiaVersion = s.version
	if status.CommitHash == "" {
		status.Containers = make([]string, 0)
		render.Render(w, r, res.Message(r, "status retrieved", http.StatusOK,
			"status", status))
		return
	}
	if err != nil {
		render.Render(w, r, res.ErrInternalServer(r, "failed to get status", err))
		return
	}

	render.Render(w, r, res.Message(r, "status retrieved", http.StatusOK,
		"status", status))
}
