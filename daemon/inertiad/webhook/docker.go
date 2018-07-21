package webhook

import "strings"

// DockerWebhook represents a push to DockerHub
// see https://docs.docker.com/docker-hub/webhooks/
type DockerWebhook struct {
	PushData dockerPushData   `json:"push_data"`
	Repo     dockerRepository `json:"repository"`
}

type dockerPushData struct {
	Pusher string `json:"pusher"`
	Tag    string `json:"tag"`
}

type dockerRepository struct {
	RepoName string `json:"repo_name"`
}

// GetPusher returns the user that pushed to DockerHub
func (d *DockerWebhook) GetPusher() string {
	return d.PushData.Pusher
}

// GetTag returns the tag that was pushed to DockerHub
func (d *DockerWebhook) GetTag() string {
	return d.PushData.Tag
}

// GetRepoName returns the full repository name
func (d *DockerWebhook) GetRepoName() string {
	return d.Repo.RepoName
}

// GetName returns the namespace
func (d *DockerWebhook) GetName() string {
	return strings.Split(d.Repo.RepoName, "/")[1]
}

// GetOwner returns the repository owner
func (d *DockerWebhook) GetOwner() string {
	return strings.Split(d.Repo.RepoName, "/")[0]
}
