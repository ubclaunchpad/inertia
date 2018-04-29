package local

import (
	"errors"

	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/common"
)

// GetDeployment returns the local deployment setup
func GetDeployment(name string) (*client.Deployment, error) {
	config, _, err := GetProjectConfigFromDisk()
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

	return &client.Deployment{
		RemoteVPS:  remote,
		Repository: repo,
		Auth:       auth,
		BuildType:  config.BuildType,
		Project:    config.Project,
	}, nil
}
