package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/gorilla/websocket"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/auth"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/cfg"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
)

var (
	// daemonVersion indicates the daemon's corresponding Inertia daemonVersion
	daemonVersion string

	// deployment is the currently deployed project on this remote
	deployment project.Deployer

	// configuration
	conf *cfg.Config

	// socketUpgrader specifies parameters for upgrading an HTTP connection to a WebSocket connection
	socketUpgrader = websocket.Upgrader{}
)

const (
	msgNoDeployment = "No deployment is currently active on this remote - try running 'inertia $REMOTE up'"
)

// run starts the daemon
func run(host, port, version string) {
	conf = cfg.New()
	daemonVersion = version

	var (
		daemonSSLCert    = path.Join(conf.SSLDirectory, "daemon.cert")
		daemonSSLKey     = path.Join(conf.SSLDirectory, "daemon.key")
		userDatabasePath = path.Join(conf.DataDirectory, "users.db")
	)

	// Download build tools
	cli, err := docker.NewEnvClient()
	if err != nil {
		println(err.Error())
		println("Failed to start Docker client - shutting down daemon.")
		return
	}
	println("Downloading build tools...")
	go downloadDeps(cli, conf.DockerComposeVersion, conf.HerokuishVersion)

	// Check if the cert files are available.
	println("Checking for existing SSL certificates in " + conf.SSLDirectory + "...")
	_, err = os.Stat(daemonSSLCert)
	certNotPresent := os.IsNotExist(err)
	_, err = os.Stat(daemonSSLKey)
	keyNotPresent := os.IsNotExist(err)

	// If they are not available, generate new ones.
	if keyNotPresent && certNotPresent {
		println("No certificates found - generating new ones...")
		err = crypto.GenerateCertificate(daemonSSLCert, daemonSSLKey, host+":"+port, "RSA")
		if err != nil {
			println(err.Error())
			return
		}
	}

	webPrefix := "/web/"
	handler, err := auth.NewPermissionsHandler(
		userDatabasePath, host, 120,
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
			webPrefix, http.FileServer(http.Dir("/daemon/inertia-web")),
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
	handler.AttachAdminRestrictedHandlerFunc("/env", envHandler)

	// Root "ok" endpoint
	handler.AttachPublicHandlerFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Serve daemon on port
	println("Serving daemon on port " + port)
	fmt.Println(http.ListenAndServeTLS(
		":"+port,
		daemonSSLCert,
		daemonSSLKey,
		handler,
	))
}

func downloadDeps(cli *docker.Client, images ...string) {
	var wait sync.WaitGroup
	wait.Add(len(images))
	for _, i := range images {
		go dockerPull(i, cli, &wait)
	}
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
