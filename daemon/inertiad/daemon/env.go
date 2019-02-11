package daemon

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-chi/render"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

// envHandler manages requests to manage environment variables
func (s *Server) envHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		envPostHandler(s, w, r)
	} else if r.Method == "GET" {
		envGetHandler(s, w, r)
	}
}

func envPostHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	// Set up logger
	stream := log.NewStreamer(log.StreamerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})

	// Parse request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		stream.Error(res.ErrBadRequest(err.Error()))
		return
	}
	defer r.Body.Close()
	var envReq api.EnvRequest
	err = json.Unmarshal(body, &envReq)
	if err != nil {
		stream.Error(res.ErrBadRequest(err.Error()))
		return
	}
	if envReq.Name == "" {
		stream.Error(res.ErrBadRequest("no variable name provided"))
		return
	}

	manager, found := s.deployment.GetDataManager()
	if !found {
		stream.Error(res.Err("no environment manager found", http.StatusPreconditionFailed))
		return
	}

	// Add, update, or remove values from storage
	if envReq.Remove {
		err = manager.RemoveEnvVariables(envReq.Name)
	} else {
		err = manager.AddEnvVariable(
			envReq.Name, envReq.Value, envReq.Encrypt,
		)
	}
	if err != nil {
		stream.Error(res.ErrInternalServer("failed to update variable", err))
		return
	}

	stream.Success(res.Msg("environment variable saved - this will be applied the next time your container is started",
		http.StatusAccepted))
}

func envGetHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	// Set up logger
	stream := log.NewStreamer(log.StreamerOptions{
		Request:    r,
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})

	manager, found := s.deployment.GetDataManager()
	if !found {
		stream.Error(res.Err("no environment manager found", http.StatusPreconditionFailed))
		return
	}

	values, err := manager.GetEnvVariables(false)
	if err != nil {
		stream.Error(res.ErrInternalServer("failed to retrieve environment variables", err))
		return
	}

	render.Render(w, r, res.Msg("configured environment variables retrieved", http.StatusOK,
		"variables", values))
}
