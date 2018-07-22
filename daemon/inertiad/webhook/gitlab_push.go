package webhook

// GitlabPushEvent represents a push to a Gitlab repository
// see https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#push-events
type gitlabPushEvent struct {
	eventType string
	ref       string
	name      string
	gitURL    string
	sshURL    string
}

func parseGitlabPushEvent(rawJSON map[string]interface{}) gitlabPushEvent {
	// Extract push details (similar to Github)
	ref := rawJSON["ref"].(string)
	repo := rawJSON["repository"].(map[string]interface{})

	name := repo["name"].(string)
	gitURL := repo["git_http_url"].(string)
	sshURL := repo["git_ssh_url"].(string)

	return gitlabPushEvent{
		eventType: PushEvent,
		ref:       ref,
		name:      name,
		gitURL:    gitURL,
		sshURL:    sshURL,
	}
}

// GetEventType returns the event type of the webhook
func (g gitlabPushEvent) GetEventType() string {
	return g.eventType
}

// GetRepoName returns the repo name
func (g gitlabPushEvent) GetRepoName() string {
	return g.name
}

// GetRef returns the full ref
func (g gitlabPushEvent) GetRef() string {
	return g.ref
}

// GetGitURL returns the git clone URL
func (g gitlabPushEvent) GetGitURL() string {
	return g.gitURL
}

// GetSSHURL returns the ssh URL
func (g gitlabPushEvent) GetSSHURL() string {
	return g.sshURL
}
