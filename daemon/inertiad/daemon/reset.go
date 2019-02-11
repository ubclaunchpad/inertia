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

	logger := log.NewLogger(log.LoggerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})
	defer logger.Close()

	// Goodbye deployment
	if err := s.deployment.Destroy(s.docker, logger); err != nil {
		logger.Error(res.ErrInternalServer("failed to remove deployment", err))
		return
	}

	logger.Success(res.MsgOK("project removed"))
}
