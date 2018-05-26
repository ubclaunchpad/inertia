package main

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/auth"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
)

var (
	// daemonVersion indicates the daemon's corresponding Inertia daemonVersion
	daemonVersion string

	// deployment is the currently deployed project on this remote
	deployment project.Deployer
)

const (
	msgNoDeployment = "No deployment is currently active on this remote - try running 'inertia $REMOTE up'"

	// specify location of SSL certificate
	sslDirectory  = "/app/host/.inertia/.ssl/"
	daemonSSLCert = sslDirectory + "daemon.cert"
	daemonSSLKey  = sslDirectory + "daemon.key"
)

// run starts the daemon
func run(host, port, version string) {
	daemonVersion = version

	// Download build tools
	cli, err := docker.NewEnvClient()
	if err != nil {
		println(err.Error())
		println("Failed to start Docker client - shutting down daemon.")
		return
	}
	println("Downloading build tools...")
	go downloadDeps(cli)

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
		err = auth.GenerateCertificate(daemonSSLCert, daemonSSLKey, host+":"+port, "RSA")
		if err != nil {
			println(err.Error())
			return
		}
	}

	webPrefix := "/web/"
	handler, err := auth.NewPermissionsHandler(
		auth.UserDatabasePath, host, 120,
	)
	if err != nil {
		println(err.Error())
		return
	}
	defer handler.Close()

	// Inertia web
	handler.AttachPublicHandler(
		webPrefix,
		http.StripPrefix(
			webPrefix, http.FileServer(http.Dir("/app/inertia-web")),
		),
	)

	// GitHub webhook endpoint
	handler.AttachPublicHandlerFunc("/webhook", gitHubWebHookHandler)

	// CLI API endpoints
	handler.AttachUserRestrictedHandlerFunc("/status", statusHandler)
	handler.AttachUserRestrictedHandlerFunc("/logs", logHandler)
	handler.AttachAdminRestrictedHandlerFunc("/up", upHandler)
	handler.AttachAdminRestrictedHandlerFunc("/down", downHandler)
	handler.AttachAdminRestrictedHandlerFunc("/reset", resetHandler)

	// Root "ok" endpoint
	handler.AttachPublicHandlerFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Serve daemon on port
	println("Serving daemon on port " + port)
	println(http.ListenAndServeTLS(
		":"+port,
		daemonSSLCert,
		daemonSSLKey,
		handler,
	))
}

func downloadDeps(cli *docker.Client) {
	var wait sync.WaitGroup
	wait.Add(2)
	go dockerPull(project.DockerComposeVersion, cli, &wait)
	go dockerPull(project.HerokuishVersion, cli, &wait)
	wait.Wait()
	cli.Close()
}

func dockerPull(image string, cli *docker.Client, wait *sync.WaitGroup) {
	defer wait.Done()
	println("Downloading " + image)
	_, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		println(err.Error())
	} else {
		println(image + " download complete")
	}
}
