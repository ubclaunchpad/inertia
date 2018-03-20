package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertia/project"
	git "gopkg.in/src-d/go-git.v4"
)

// upHandler tries to bring the deployment online
func upHandler(w http.ResponseWriter, r *http.Request) {
	println("UP request received")

	// Get github URL from up request
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
	logger := newLogger(upReq.Stream, w)
	gitOpts := upReq.GitOptions
	defer logger.Close()

	// Check for existing git repository, clone if no git repository exists.
	err = common.CheckForGit(project.Directory)
	if err != nil {
		logger.Println("No git repository present.")
		err = setUpProject(gitOpts.RemoteURL, gitOpts.Branch, logger.GetWriter())
		if err != nil {
			logger.Err(err.Error(), http.StatusPreconditionFailed)
			return
		}
	}

	repo, err := git.PlainOpen(project.Directory)
	if err != nil {
		logger.Err(err.Error(), http.StatusPreconditionFailed)
		return
	}

	// Check for matching remotes
	err = common.CompareRemotes(repo, gitOpts.RemoteURL)
	if err != nil {
		logger.Err(err.Error(), http.StatusPreconditionFailed)
		return
	}

	// Update and deploy project
	cli, err := docker.NewEnvClient()
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	err = project.Deploy(repo, gitOpts.Branch, cli, logger.GetWriter())
	if err != nil {
		logger.Err(err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Success("Project startup initiated!", http.StatusCreated)
}
