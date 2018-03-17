package daemon

import (
	"context"
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"

	"github.com/ubclaunchpad/inertia/common"
)

const (
	// specify location of deployed project
	projectDirectory = "/app/host/project"

	// specify location of SSL certificate
	sslDirectory = "/app/ssl/"

	// specify docker-compose version
	dockerCompose = "docker/compose:1.18.0"

	// specify common responses here
	noContainersResp            = "There are currently no active containers."
	malformedAuthStringErrorMsg = "Malformed authentication string"
	tokenInvalidErrorMsg        = "Token invalid"

	defaultSecret = "inertia"
)

// daemonVersion indicates the daemon's corresponding Inertia daemonVersion
var daemonVersion string

// Run starts the daemon
func Run(port, version string) {
	daemonVersion = version

	// Download docker-compose image
	println("Downloading docker-compose...")
	cli, err := docker.NewEnvClient()
	if err != nil {
		println(err.Error())
		println("Failed to start Docker client - shutting down daemon.")
		return
	}
	_, err = cli.ImagePull(context.Background(), dockerCompose, types.ImagePullOptions{})
	if err != nil {
		println(err.Error())
		println("Failed to pull docker-compose image - shutting down daemon.")
		cli.Close()
		return
	}
	cli.Close()

	// Run daemon on port
	println("Serving daemon on port " + port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", gitHubWebHookHandler)
	mux.HandleFunc("/up", authorized(upHandler, GetAPIPrivateKey))
	mux.HandleFunc("/down", authorized(downHandler, GetAPIPrivateKey))
	mux.HandleFunc("/status", authorized(statusHandler, GetAPIPrivateKey))
	mux.HandleFunc("/reset", authorized(resetHandler, GetAPIPrivateKey))
	mux.HandleFunc("/logs", authorized(logHandler, GetAPIPrivateKey))
	mux.HandleFunc("/health-check", authorized(healthCheckHandler, GetAPIPrivateKey))
	print(http.ListenAndServeTLS(
		":"+port,
		"./server.cert",
		"./server.key",
		mux,
	))
}

// healthCheckHandler returns a 200 if the daemon is happy.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, common.DaemonOkResp)
}
