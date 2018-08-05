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
	"github.com/ubclaunchpad/inertia/daemon/inertiad/build"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/cfg"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
)

const (
	msgNoDeployment = "No deployment is currently active on this remote - try running 'inertia [remote] up'"
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

// run starts the daemon
func run(host, port, version string) {
	// Load config and set globals
	conf = cfg.New()
	daemonVersion = version

	// Generate paths
	var (
		daemonSSLCert = path.Join(conf.SSLDirectory, "daemon.cert")
		daemonSSLKey  = path.Join(conf.SSLDirectory, "daemon.key")

		userDatabasePath    = path.Join(conf.DataDirectory, "users.db")
		projectDatabasePath = path.Join(conf.DataDirectory, "project.db")
	)

	// Download build tools
	cli, err := containers.NewDockerClient()
	if err != nil {
		println(err.Error())
		println("Failed to start Docker client - shutting down daemon.")
		return
	}
	go downloadDeps(cli, conf.DockerComposeVersion, conf.HerokuishVersion)

	// Check if the cert files are available.
	_, err = os.Stat(daemonSSLCert)
	certNotPresent := os.IsNotExist(err)
	_, err = os.Stat(daemonSSLKey)
	keyNotPresent := os.IsNotExist(err)

	// If they are not available, generate new ones.
	if keyNotPresent && certNotPresent {
		fmt.Printf("No certificates found in %s - generating new ones...", conf.SSLDirectory)
		err = crypto.GenerateCertificate(daemonSSLCert, daemonSSLKey, host+":"+port, "RSA")
		if err != nil {
			println(err.Error())
			return
		}
	} else {
		fmt.Printf("Found certificates in %s (%s, %s)", conf.SSLDirectory, daemonSSLCert, daemonSSLKey)
	}

	// Set up deployment
	deployment, err = project.NewDeployment(
		conf.ProjectDirectory, projectDatabasePath,
		build.NewBuilder(*conf, containers.StopActiveContainers))
	if err != nil {
		println(err.Error())
		return
	}
	println("Deployment manager successfully created")

	// Watch container events
	go func() {
		logsCh, errCh := deployment.Watch(cli)
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
	webPrefix := "/web/"
	handler, err := auth.NewPermissionsHandler(
		userDatabasePath, host, 120,
	)
	if err != nil {
		println(err.Error())
		return
	}
	defer handler.Close()
	println("Permissions manager successfully created")

	// Inertia web
	handler.AttachPublicHandler(
		webPrefix,
		http.StripPrefix(
			webPrefix, http.FileServer(http.Dir("/daemon/inertia-web")),
		),
	)

	// GitHub webhook endpoint
	handler.AttachPublicHandlerFunc("/webhook", webhookHandler)

	// CLI API endpoints
	handler.AttachUserRestrictedHandlerFunc("/status", statusHandler)
	handler.AttachUserRestrictedHandlerFunc("/logs", logHandler)
	handler.AttachAdminRestrictedHandlerFunc("/up", upHandler)
	handler.AttachAdminRestrictedHandlerFunc("/down", downHandler)
	handler.AttachAdminRestrictedHandlerFunc("/reset", resetHandler)
	handler.AttachAdminRestrictedHandlerFunc("/env", envHandler)
	handler.AttachAdminRestrictedHandlerFunc("/prune", pruneHandler)

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
