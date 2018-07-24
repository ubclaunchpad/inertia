package webhook

import "strings"

// DockerWebhook represents a push to DockerHub
// see https://docs.docker.com/docker-hub/webhooks/
type DockerWebhook struct {
	pusher   string
	tag      string
	repoName string
}

// Extract DockerHub push details
func parseDocker(rawJSON map[string]interface{}) (*DockerWebhook, error) {
	pushData := rawJSON["push_data"].(map[string]interface{})
	repo := rawJSON["repository"].(map[string]interface{})

	pusher := pushData["pusher"].(string)
	tag := pushData["tag"].(string)
	repoName := repo["repo_name"].(string)

	payload := &DockerWebhook{
		pusher:   pusher,
		tag:      tag,
		repoName: repoName,
	}
	return payload, nil
}

// GetPusher returns the user that pushed to DockerHub
func (d *DockerWebhook) GetPusher() string {
	return d.pusher
}

// GetTag returns the tag that was pushed to DockerHub
func (d *DockerWebhook) GetTag() string {
	return d.tag
}

// GetRepoName returns the full repository name
func (d *DockerWebhook) GetRepoName() string {
	return d.repoName
}

// GetName returns the namespace
func (d *DockerWebhook) GetName() string {
	return strings.Split(d.repoName, "/")[1]
}

// GetOwner returns the repository owner
func (d *DockerWebhook) GetOwner() string {
	return strings.Split(d.repoName, "/")[0]
}
