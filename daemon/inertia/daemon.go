package main

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/daemon/inertia/auth"
	"github.com/ubclaunchpad/inertia/daemon/inertia/project"
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
	sslDirectory  = "/app/host/ssl/"
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

	// GitHub webhook endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/", gitHubWebHookHandler)

	// Inertia web - PermissionsHandler is used to authenticate web
	// app access and manage users
	webPrefix := "/web/"
	permHandler, err := auth.NewPermissionsHandler(
		auth.UserDatabasePath, host, webPrefix, 120,
	)
	if err != nil {
		println(err.Error())
		return
	}
	defer permHandler.Close()
	permHandler.AttachPublicHandler(
		"/", http.FileServer(http.Dir("/app/inertia-web")),
	)
	mux.Handle(webPrefix, http.StripPrefix(webPrefix, permHandler))

	// CLI API endpoints
	mux.HandleFunc("/up", auth.Authorized(upHandler, auth.GetAPIPrivateKey))
	mux.HandleFunc("/down", auth.Authorized(downHandler, auth.GetAPIPrivateKey))
	mux.HandleFunc("/status", auth.Authorized(statusHandler, auth.GetAPIPrivateKey))
	mux.HandleFunc("/reset", auth.Authorized(resetHandler, auth.GetAPIPrivateKey))
	mux.HandleFunc("/logs", auth.Authorized(logHandler, auth.GetAPIPrivateKey))

	// Serve daemon on port
	println("Serving daemon on port " + port)
	println(http.ListenAndServeTLS(
		":"+port,
		daemonSSLCert,
		daemonSSLKey,
		mux,
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
