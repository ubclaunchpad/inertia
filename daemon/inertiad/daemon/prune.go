package daemon

import (
	"net/http"
	"os"

	"github.com/go-chi/render"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

// pruneHandler cleans up Docker assets
func (s *Server) pruneHandler(w http.ResponseWriter, r *http.Request) {
	if s.deployment == nil {
		render.Render(w, r, res.Err(msgNoDeployment, http.StatusPreconditionFailed))
		return
	}

	logger := log.NewLogger(log.LoggerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})
	defer logger.Close()

	if err := s.deployment.Prune(s.docker, logger); err != nil {
		logger.Error(res.ErrInternalServer("failed to prune Docker assets", err))
		return
	}

	logger.Success(res.MsgOK("docker assets have been pruned"))
}
