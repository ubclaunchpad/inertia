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

type Server struct {
	version string

	deployment project.Deployer
	state      cfg.Config

	docker    *docker.Client
	websocket *websocket.Upgrader
}

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

func (s *Server) Run(host, port string) error {
	var (
		err  error
		cert = path.Join(s.state.SSLDirectory, "daemon.cert")
		key  = path.Join(s.state.SSLDirectory, "daemon.key")
	)

	// Check if the cert files are available.
	_, err = os.Stat(cert)
	certNotPresent := os.IsNotExist(err)
	_, err = os.Stat(key)
	keyNotPresent := os.IsNotExist(err)

	// If they are not available, generate new ones.
	if keyNotPresent && certNotPresent {
		fmt.Printf("No certificates found in %s - generating new ones...", s.state.SSLDirectory)
		if err = crypto.GenerateCertificate(cert, key, host+":"+port, "RSA"); err != nil {
			return err
		}
	} else {
		fmt.Printf("Found certificates in %s (%s, %s)",
			s.state.SSLDirectory, cert, key)
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
	var (
		webPrefix        = "/web/"
		userDatabasePath = path.Join(s.state.DataDirectory, "users.db")
	)
	handler, err := auth.NewPermissionsHandler(
		userDatabasePath, host, 120)
	if err != nil {
		return err
	}
	defer handler.Close()
	println("Permissions manager successfully created")

	// Inertia web
	handler.AttachPublicHandler(
		webPrefix,
		http.StripPrefix(webPrefix, http.FileServer(http.Dir("/daemon/inertia-web"))))

	// GitHub webhook endpoint
	handler.AttachPublicHandlerFunc("/webhook", s.webhookHandler)

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
		s.pruneHandler, http.MethodGet)
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

func (s *Server) Close() {
	s.deployment.Down(s.docker, os.Stdout)
	s.docker.Close()
}
