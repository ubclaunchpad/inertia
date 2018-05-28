package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	docker "github.com/docker/docker/client"
	"github.com/gorilla/websocket"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
)

// logHandler handles requests for container logs
func logHandler(w http.ResponseWriter, r *http.Request) {
	var (
		logger *log.DaemonLogger
		socket *websocket.Conn
	)

	// Get container name and stream from request query params
	params := r.URL.Query()
	container := params.Get("container")
	stream, err := strconv.ParseBool(params.Get("stream"))
	if err != nil {
		println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Upgrade to websocket connection if required
	if stream {
		socket, err := socketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger = log.NewLogger(os.Stdout, socket, w)
	} else {
		logger = log.NewLogger(os.Stdout, nil, w)
	}
	defer logger.Close()

	if !strings.Contains(container, "inertia-daemon") && deployment == nil {
		logger.WriteErr(msgNoDeployment, http.StatusPreconditionFailed)
		return
	}

	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	logs, err := project.ContainerLogs(cli, project.LogOptions{
		Container: container,
		Stream:    stream,
	})
	if err != nil {
		if docker.IsErrContainerNotFound(err) {
			logger.WriteErr(err.Error(), http.StatusNotFound)
		} else {
			logger.WriteErr(err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer logs.Close()

	if stream {
		stop := make(chan struct{})
		log.FlushRoutine(log.NewWebSocketTextWriter(socket), logs, stop)
		defer close(stop)
	} else {
		buf := new(bytes.Buffer)
		buf.ReadFrom(logs)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, buf.String())
	}
}
