package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"

	"github.com/ubclaunchpad/inertia/common"
)

// daemonVersion indicates the daemon's corresponding Inertia daemonVersion
var daemonVersion string

const (
	// specify location of deployed project
	projectDirectory = "/app/host/project"

	// specify location of SSL certificate
	sslDirectory  = "/app/host/ssl/"
	daemonSSLCert = sslDirectory + "daemon.cert"
	daemonSSLKey  = sslDirectory + "daemon.key"

	// specify docker-compose version
	dockerCompose = "docker/compose:1.18.0"

	// specify common responses here
	noContainersResp            = "There are currently no active containers."
	malformedAuthStringErrorMsg = "Malformed authentication string"
	tokenInvalidErrorMsg        = "Token invalid"

	defaultSecret = "inertia"
)

// run starts the daemon
func run(host, port, version string) {
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

	// Check if the cert files are available.
	println("Checking for existing SSL certificates in " + sslDirectory + "...")
	_, err = os.Stat(daemonSSLCert)
	certNotPresent := os.IsNotExist(err)
	_, err = os.Stat(daemonSSLKey)
	keyNotPresent := os.IsNotExist(err)
	sslRequirementsPresent := !(keyNotPresent && certNotPresent)

	// If they are not available, generate new ones.
	if !sslRequirementsPresent {
		println("No certificates found - generating new ones...")
		err = generateCertificate(daemonSSLCert, daemonSSLKey, host+":"+port)
		if err != nil {
			println(err.Error())
			return
		}
	}

	// API endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("/up", authorized(upHandler, getAPIPrivateKey))
	mux.HandleFunc("/down", authorized(downHandler, getAPIPrivateKey))
	mux.HandleFunc("/status", authorized(statusHandler, getAPIPrivateKey))
	mux.HandleFunc("/reset", authorized(resetHandler, getAPIPrivateKey))
	mux.HandleFunc("/logs", authorized(logHandler, getAPIPrivateKey))
	mux.HandleFunc("/health-check", authorized(healthCheckHandler, getAPIPrivateKey))

	// GitHub webhook endpoint
	mux.HandleFunc("/", gitHubWebHookHandler)

	// Inertia web
	mux.Handle("/admin/", http.FileServer(http.Dir("./inertia-web")))

	// Serve daemon on port
	println("Serving daemon on port " + port)
	println(http.ListenAndServeTLS(
		":"+port,
		daemonSSLCert,
		daemonSSLKey,
		mux,
	))
}

// healthCheckHandler returns a 200 if the daemon is happy.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, common.DaemonOkResp)
}
