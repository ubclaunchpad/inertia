package daemon

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
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
	logger := log.NewLogger(log.LoggerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})
	// Parse request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var envReq api.EnvRequest
	err = json.Unmarshal(body, &envReq)
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusBadRequest)
		return
	}
	if envReq.Name == "" {
		logger.WriteErr("no variable name provided", http.StatusBadRequest)
	}

	manager, found := s.deployment.GetDataManager()
	if !found {
		logger.WriteErr("no environment manager found", http.StatusPreconditionFailed)
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
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WriteSuccess("environment variable saved - this will be applied the next time your container is started", http.StatusAccepted)
}

func envGetHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	// Set up logger
	logger := log.NewLogger(log.LoggerOptions{
		Stdout:     os.Stdout,
		HTTPWriter: w,
	})

	manager, found := s.deployment.GetDataManager()
	if !found {
		logger.WriteErr("no environment manager found", http.StatusPreconditionFailed)
		return
	}

	values, err := manager.GetEnvVariables(false)
	if err != nil {
		logger.WriteErr(err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(values)
}
