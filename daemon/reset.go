package daemon

import (
	"fmt"
	"net/http"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
)

// resetHandler shuts down and wipes the project directory
func resetHandler(w http.ResponseWriter, r *http.Request) {
	println("RESET request received")

	logger := newLogger(false, w)
	defer logger.Close()

	cli, err := docker.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	err = killActiveContainers(cli, logger.GetWriter())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = common.RemoveContents(projectDirectory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Project removed from remote.")
}
