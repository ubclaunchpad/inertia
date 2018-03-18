package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ubclaunchpad/inertia/common"
	git "gopkg.in/src-d/go-git.v4"
	//"io/ioutil"
)

// Deployment manages a deployment and implements the
// DaemonRequester interface.
type Deployment struct {
	*RemoteVPS
	Repository *git.Repository
	Auth       string
	Project    string
}

// GetDeployment returns the local deployment setup
// TODO: add args to support getting the appropriate deployment
// based on the command (aka remote) that calls it
func GetDeployment(name string) (*Deployment, error) {
	config, err := GetProjectConfigFromDisk()
	if err != nil {
		return nil, err
	}

	repo, err := common.GetLocalRepo()
	if err != nil {
		return nil, err
	}

	remote, found := config.GetRemote(name)
	if !found {
		return nil, errors.New("Remote not found")
	}
	auth := remote.Daemon.Token

	return &Deployment{
		RemoteVPS:  remote,
		Repository: repo,
		Auth:       auth,
		Project:    config.Project,
	}, nil
}

// Up brings the project up on the remote VPS instance specified
// in the deployment object.
func (d *Deployment) Up(project string, stream bool) (*http.Response, error) {
	host := "http://" + d.RemoteVPS.GetIPAndPort() + "/up"

	// TODO: Support other repo names.
	origin, err := d.Repository.Remote("origin")
	if err != nil {
		return nil, err
	}

	reqContent := common.DaemonRequest{
		Stream: stream,
		Project: project,
		GitOptions: &common.GitOptions{
			RemoteURL: common.GetSSHRemoteURL(origin.Config().URLs[0]),
			Branch:    d.Branch,
		},
	}
	body, err := json.Marshal(reqContent)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", host, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.Auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("Error when deploying project: " + err.Error())
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

// Logs get logs
func (d *Deployment) Logs(stream bool, container string) (*http.Response, error) {
	host := "http://" + d.RemoteVPS.GetIPAndPort() + "/logs"

	reqContent := common.DaemonRequest{
		Stream:    stream,
		Container: container,
	}
	body, err := json.Marshal(reqContent)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", host, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.Auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("Error when deploying project: " + err.Error())
	}
	return resp, nil
}
