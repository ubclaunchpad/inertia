package daemon

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/gorilla/websocket"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/auth"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/cfg"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
)

// Server is the core component of Inertiad, and hosts its API and deployment manager
type Server struct {
	version string

	deployment project.Deployer
	state      cfg.Config

	docker    *docker.Client
	websocket *websocket.Upgrader
}

// New instantiates a new Inertiad server
func New(version string, state cfg.Config, deployment project.Deployer) (*Server, error) {
	// Establish connection with dockerd
	cli, err := containers.NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to start Docker client: %s", err.Error())
	}

	// Download build tools
	go downloadDeps(cli, state.DockerComposeVersion)

	return &Server{
		version: version,

		deployment: deployment,
		state:      state,

		docker: cli,
		websocket: &websocket.Upgrader{
			HandshakeTimeout: 5 * time.Second,
		},
	}, nil
}

// Run starts the server
func (s *Server) Run(host, port string) error {
	var (
		err    error
		sslDir = path.Join(s.state.SecretsDirectory, "ssl")
		cert   = path.Join(sslDir, "daemon.cert")
		key    = path.Join(sslDir, "daemon.key")
	)

	// Check if the cert files are available.
	_, err = os.Stat(cert)
	certNotPresent := os.IsNotExist(err)
	_, err = os.Stat(key)
	keyNotPresent := os.IsNotExist(err)

	// If they are not available, generate new ones.
	if keyNotPresent && certNotPresent {
		fmt.Printf("No certificates found in %s - generating new ones...", sslDir)
		if err = crypto.GenerateCertificate(cert, key, host+":"+port, "RSA"); err != nil {
			return err
		}
	} else {
		fmt.Printf("Found certificates in %s (%s, %s)",
			sslDir, cert, key)
	}

	// Watch container events
	go func() {
		logsCh, errCh := s.deployment.Watch(s.docker)
		for {
			select {
			case err := <-errCh:
				if err != nil {
					println(err.Error())
					return
				}
			case event := <-logsCh:
				println(event)
			}
		}
	}()

	// Set up endpoints
	handler, err := auth.NewPermissionsHandler(path.Join(s.state.DataDirectory, "users.db"), host, 120)
	if err != nil {
		return err
	}
	defer handler.Close()
	println("Permissions manager successfully created")

	// GitHub webhook endpoint
	handler.AttachPublicHandlerFunc("/webhook",
		s.webhookHandler, http.MethodPost)

	// API endpoints
	handler.AttachUserRestrictedHandlerFunc("/status",
		s.statusHandler, http.MethodGet)
	handler.AttachUserRestrictedHandlerFunc("/logs",
		s.logHandler, http.MethodGet)
	handler.AttachAdminRestrictedHandlerFunc("/up",
		s.upHandler, http.MethodPost)
	handler.AttachAdminRestrictedHandlerFunc("/down",
		s.downHandler, http.MethodPost)
	handler.AttachAdminRestrictedHandlerFunc("/reset",
		s.resetHandler, http.MethodPost)
	handler.AttachAdminRestrictedHandlerFunc("/env",
		s.envHandler, http.MethodGet, http.MethodPost)
	handler.AttachAdminRestrictedHandlerFunc("/prune",
		s.pruneHandler, http.MethodPost)
	handler.AttachAdminRestrictedHandlerFunc("/token",
		tokenHandler, http.MethodGet)

	// Root "ok" endpoint
	handler.AttachPublicHandlerFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Serve daemon on port
	println("Serving daemon on port " + port)
	return http.ListenAndServeTLS(
		":"+port,
		cert,
		key,
		handler)
}

// Close releases server assets
func (s *Server) Close() {
	s.deployment.Down(s.docker, os.Stdout)
	s.docker.Close()
}
