package daemon

import (
	"net/http"
	"os"

	"github.com/go-chi/render"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

const (
	msgNoDeployment = "No deployment is currently active on this remote - try running 'inertia [remote] up'"
)

// downHandler tries to take the deployment offline
func (s *Server) downHandler(w http.ResponseWriter, r *http.Request) {
	if status, _ := s.deployment.GetStatus(s.docker); len(status.Containers) == 0 {
		render.Render(w, r, res.Err(msgNoDeployment, http.StatusPreconditionFailed))
		return
	}

	logger := log.NewLogger(log.LoggerOptions{
		Request:    r,
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})
	defer logger.Close()

	if err := s.deployment.Down(s.docker, logger); err == containers.ErrNoContainers {
		logger.Error(res.Err(err.Error(), http.StatusPreconditionFailed))
		return
	} else if err != nil {
		logger.Error(res.ErrInternalServer("failed to shut down project", err))
		return
	}

	logger.Success(res.MsgOK("project shut down"))
}
