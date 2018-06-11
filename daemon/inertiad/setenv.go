package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
)

// setenvHandler manages requests to set environment variables
func setenvHandler(w http.ResponseWriter, r *http.Request) {
	// Set up logger
	logger := log.NewLogger(log.LoggerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})
	if deployment == nil {
		logger.WriteErr(msgNoDeployment, http.StatusPreconditionFailed)
		return
	}

	// Get container name and stream from request query params
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var envReq common.EnvRequest
	err = json.Unmarshal(body, &envReq)
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusBadRequest)
		return
	}
	if envReq.Name == "" {
		logger.WriteErr("no variable name provided", http.StatusBadRequest)
	}

	// Add, update, or remove values from storage
	if envReq.Remove {
		err = deployment.GetDataManager().RemoveEnvVariable(envReq.Name)
	} else {
		err = deployment.GetDataManager().AddEnvVariable(
			envReq.Name, envReq.Value, envReq.Encrypt,
		)
	}
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	// Update values in containers
	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	err = deployment.UpdateContainerEnvironmentValues(cli)
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WriteSuccess("environment variable updated", http.StatusAccepted)
}
