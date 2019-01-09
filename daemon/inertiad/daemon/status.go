package daemon

import (
	"encoding/json"
	"net/http"

	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
)

// statusHandler returns a formatted string about the status of the
// deployment and lists currently active project containers
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	cli, err := containers.NewDockerClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	status, err := s.deployment.GetStatus(cli)
	if status.CommitHash == "" {
		status := &common.DeploymentStatus{
			InertiaVersion: s.version,
			Containers:     make([]string, 0),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(status)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status.InertiaVersion = s.version

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
