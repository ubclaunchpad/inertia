package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
)

// logHandler handles requests for container logs
func logHandler(w http.ResponseWriter, r *http.Request) {
	var (
		logger *log.DaemonLogger
		stream bool
	)

	// Get container name and stream from request query params
	params := r.URL.Query()
	container := params.Get("container")
	streamParam := params.Get("stream")
	if streamParam != "" {
		s, err := strconv.ParseBool(streamParam)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		stream = s
	} else {
		stream = false
	}

	// Upgrade to websocket connection if required, otherwise just set up a
	// standard logger
	if stream {
		socket, err := socketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger = log.NewLogger(log.LoggerOptions{
			Stdout:     os.Stdout,
			Socket:     socket,
			HTTPWriter: w,
		})
	} else {
		logger = log.NewLogger(log.LoggerOptions{
			Stdout:     os.Stdout,
			HTTPWriter: w,
		})
	}
	defer logger.Close()

	// If no deployment is online, error unless the client is requesting for
	// the daemon's logs
	if deployment == nil && !strings.Contains(container, "inertia-daemon") {
		logger.WriteErr(msgNoDeployment, http.StatusPreconditionFailed)
		return
	}

	cli, err := containers.NewDockerClient()
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	logs, err := containers.ContainerLogs(cli, containers.LogOptions{
		Container: container,
		Stream:    stream,
	})
	if err != nil {
		if docker.IsErrNotFound(err) {
			logger.WriteErr(err.Error(), http.StatusNotFound)
		} else {
			logger.WriteErr(err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer logs.Close()

	if stream {
		stop := make(chan struct{})
		socket, err := logger.GetSocketWriter()
		if err != nil {
			logger.WriteErr(err.Error(), http.StatusInternalServerError)
		}
		log.FlushRoutine(socket, logs, stop)
		defer close(stop)
	} else {
		buf := new(bytes.Buffer)
		buf.ReadFrom(logs)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, buf.String())
	}
}
