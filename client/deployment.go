package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/ubclaunchpad/inertia/common"
	git "gopkg.in/src-d/go-git.v4"
)

// Deployment manages a deployment
type Deployment struct {
	*RemoteVPS
	Repository *git.Repository
	Auth       string
	Project	   string
}

// GetDeployment returns the local deployment setup
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
	// TODO: Support other Git remotes.
	origin, err := d.Repository.Remote("origin")
	if err != nil {
		return nil, err
	}

	reqContent := &common.DaemonRequest{
		Stream: stream,
		Project: project,
		Secret: d.RemoteVPS.Daemon.Secret
		GitOptions: &common.GitOptions{
			RemoteURL: common.GetSSHRemoteURL(origin.Config().URLs[0]),
			Branch:    d.Branch,
		},
	}
	return d.post("/up", reqContent)
}

// Down brings the project down on the remote VPS instance specified
// in the configuration object.
func (d *Deployment) Down() (*http.Response, error) {
	return d.post("/down", nil)
}

// Status lists the currently active containers on the remote VPS instance
func (d *Deployment) Status() (*http.Response, error) {
	return d.post("/status", nil)
}

// Reset shuts down deployment and deletes the contents of the deployment's
// project directory
func (d *Deployment) Reset() (*http.Response, error) {
	return d.post("/reset", nil)
}

// Logs get logs of given container
func (d *Deployment) Logs(stream bool, container string) (*http.Response, error) {
	reqContent := &common.DaemonRequest{
		Stream:    stream,
		Container: container,
	}
	return d.post("/logs", reqContent)
}

func (d *Deployment) post(endpoint string, requestBody *common.DaemonRequest) (*http.Response, error) {
	// Assemble URL
	url, err := url.Parse("https://" + d.RemoteVPS.GetIPAndPort())
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, endpoint)
	urlString := url.String()

	// Assemble request
	var payload io.Reader
	if requestBody != nil {
		body, err := json.Marshal(*requestBody)
		if err != nil {
			return nil, err
		}
		payload = bytes.NewReader(body)
	} else {
		payload = nil
	}
	req, err := http.NewRequest("POST", urlString, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.Auth)

	// Make HTTPS request
	tr := &http.Transport{
		// Our certificates are self-signed, so will raise
		// a warning - currently, we ask our client to ignore
		// this warning.
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	return client.Do(req)
}
