package daemon

import (
	"net/http"
	"os"

	"github.com/go-chi/render"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

// pruneHandler cleans up Docker assets
func (s *Server) pruneHandler(w http.ResponseWriter, r *http.Request) {
	if s.deployment == nil {
		render.Render(w, r, res.Err(r, msgNoDeployment, http.StatusPreconditionFailed))
		return
	}

	logger := log.NewLogger(log.LoggerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})
	defer logger.Close()

	cli, err := containers.NewDockerClient()
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	if err = s.deployment.Prune(cli, logger); err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WriteSuccess("Docker assets have been pruned.", http.StatusOK)
}
