package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertia/project"
)

// logHandler handles requests for container logs
func logHandler(w http.ResponseWriter, r *http.Request) {
	// Get container name from request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var upReq common.DaemonRequest
	err = json.Unmarshal(body, &upReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	container := upReq.Container
	logger := newLogger(upReq.Stream, w)
	defer logger.Close()

	if container != "/inertia-daemon" && deployment == nil {
		logger.Err(msgNoDeployment, http.StatusPreconditionFailed)
		return
	}

	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	logs, err := deployment.Logs(cli, project.LogOptions{
		Container: upReq.Container,
		Stream:    upReq.Stream,
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

	if upReq.Stream {
		stop := make(chan struct{})
		common.FlushRoutine(w, logs, stop)
	} else {
		buf := new(bytes.Buffer)
		buf.ReadFrom(logs)
		logger.Success(buf.String(), http.StatusOK)
	}
}
