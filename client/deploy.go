// Copyright © 2017 UBC Launch Pad team@ubclaunchpad.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ubclaunchpad/inertia/common"
	git "gopkg.in/src-d/go-git.v4"
)

// Deployment manages a deployment and implements the
// DaemonRequester interface.
type Deployment struct {
	*RemoteVPS
	Repository *git.Repository
	Auth       string
}

// DaemonRequester can make HTTP requests to the daemon.
type DaemonRequester interface {
	Up() (*http.Response, error)
	Down() (*http.Response, error)
	Status() (*http.Response, error)
	Reset() (*http.Response, error)
}

// GetDeployment returns the local deployment setup
// TODO: add args to support getting the appropriate deployment
// based on the command (aka remote) that calls it
func GetDeployment() (*Deployment, error) {
	config, err := GetProjectConfigFromDisk()
	if err != nil {
		return nil, err
	}

	repo, err := common.GetLocalRepo()
	if err != nil {
		return nil, err
	}

	auth := config.DaemonAPIToken

	return &Deployment{
		RemoteVPS:  config.CurrentRemoteVPS,
		Repository: repo,
		Auth:       auth,
	}, nil
}

// Up brings the project up on the remote VPS instance specified
// in the deployment object.
func (d *Deployment) Up() (*http.Response, error) {
	host := "http://" + d.RemoteVPS.GetIPAndPort() + "/up"

	// TODO: Support other repo names.
	origin, err := d.Repository.Remote("origin")
	if err != nil {
		return nil, err
	}

	reqContent := common.UpRequest{
		Repo: common.GetSSHRemoteURL(origin.Config().URLs[0]),
	}
	body, err := json.Marshal(reqContent)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", host, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+d.Auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("Error when deploying project")
	}
	return resp, nil
}

// Down brings the project down on the remote VPS instance specified
// in the configuration object.
func (d *Deployment) Down() (*http.Response, error) {
	host := "http://" + d.RemoteVPS.GetIPAndPort() + "/down"

	req, err := http.NewRequest("POST", host, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+d.Auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("Error when shutting down project")
	}

	return resp, nil
}

// Status lists the currently active containers on the remote VPS instance
func (d *Deployment) Status() (*http.Response, error) {
	host := "http://" + d.RemoteVPS.GetIPAndPort() + "/status"

	req, err := http.NewRequest("POST", host, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+d.Auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("Error when checking project status")
	}

	return resp, nil
}

// Reset shuts down deployment and deletes the contents of the deployment's
// project directory
func (d *Deployment) Reset() (*http.Response, error) {
	host := "http://" + d.RemoteVPS.GetIPAndPort() + "/reset"

	req, err := http.NewRequest("POST", host, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+d.Auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("Error when reseting project on deployment")
	}

	return resp, nil
}
