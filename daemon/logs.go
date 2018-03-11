package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
)

// logHandler handles requests for container logs
func logHandler(w http.ResponseWriter, r *http.Request) {
	println("LOG request received")

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

	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	ctx := context.Background()
	logs, err := cli.ContainerLogs(ctx, container, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     upReq.Stream,
		Timestamps: true,
	})
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}
	defer logs.Close()

	if upReq.Stream {
		common.FlushRoutine(w, logs)
	} else {
		buf := new(bytes.Buffer)
		buf.ReadFrom(logs)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, buf.String())
	}
}
