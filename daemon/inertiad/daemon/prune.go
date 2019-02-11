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

	var stream = log.NewStreamer(log.StreamerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})
	defer stream.Close()

	if err := s.deployment.Prune(s.docker, stream); err != nil {
		stream.Error(res.ErrInternalServer("failed to prune Docker assets", err))
		return
	}

	stream.Success(res.MsgOK("docker assets have been pruned"))
}
