package daemon

import (
	"net/http"
	"os"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
)

// resetHandler shuts down and wipes the project directory
func (s *Server) resetHandler(w http.ResponseWriter, r *http.Request) {
	if s.deployment == nil {
		http.Error(w, msgNoDeployment, http.StatusPreconditionFailed)
		return
	}

	logger := log.NewLogger(log.LoggerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})
	defer logger.Close()

	// Goodbye deployment
	if err := s.deployment.Destroy(s.docker, logger); err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WriteSuccess("Project removed from remote.", http.StatusOK)
}
