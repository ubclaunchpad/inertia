package daemon

import (
	"net/http"
	"os"

	"github.com/go-chi/render"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

// resetHandler shuts down and wipes the project directory
func (s *Server) resetHandler(w http.ResponseWriter, r *http.Request) {
	if s.deployment == nil {
		render.Render(w, r, res.Err(msgNoDeployment, http.StatusPreconditionFailed))
		return
	}

	var stream = log.NewStreamer(log.StreamerOptions{
		Request:    r,
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})
	defer stream.Close()

	// Goodbye deployment
	if err := s.deployment.Destroy(s.docker, stream); err != nil {
		stream.Error(res.ErrInternalServer("failed to remove deployment", err))
		return
	}

	stream.Success(res.MsgOK("project removed"))
}
