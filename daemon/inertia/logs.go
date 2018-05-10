package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertia/project"
	"strconv"
)

// logHandler handles requests for container logs
func logHandler(w http.ResponseWriter, r *http.Request) {
	// Get container name and stream from request query params
	//var upReq common.DaemonRequest
	q := r.URL.Query()

	container := q.Get("Container")
	stream, err := strconv.ParseBool(q.Get("Stream"))
	if err != nil {
		println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger := newLogger(stream, w)
	defer logger.Close()

	if !strings.Contains(container, "inertia-daemon") && deployment == nil {
		logger.Err(msgNoDeployment, http.StatusPreconditionFailed)
		return
	}

	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	logs, err := project.ContainerLogs(cli, project.LogOptions{
		Container: container,
		Stream:    stream,
	})
	if err != nil {
		if docker.IsErrContainerNotFound(err) {
			logger.Err(err.Error(), http.StatusNotFound)
		} else {
			logger.Err(err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer logs.Close()

	if stream {
		stop := make(chan struct{})
		common.FlushRoutine(w, logs, stop)
		defer close(stop)
	} else {
		buf := new(bytes.Buffer)
		buf.ReadFrom(logs)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, buf.String())
	}
}
