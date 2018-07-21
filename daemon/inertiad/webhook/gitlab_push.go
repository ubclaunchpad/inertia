package webhook

// GitlabPushEvent represents a push to a Gitlab repository
// see https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#push-events
type gitlabPushEvent struct {
	eventType string
	Ref       string                    `json:"ref"`
	Repo      gitlabPushEventRepository `json:"repository"`
}

// GitlabPushEventRepository represents the repository object in a Gitlab PushEvent
// see https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#push-events
type gitlabPushEventRepository struct {
	Name   string `json:"name"`
	GitURL string `json:"git_http_url"`
	SSHURL string `json:"git_ssh_url"`
}

// GetEventType returns the event type of the webhook
func (g gitlabPushEvent) GetEventType() string {
	return g.eventType
}

// GetRepoName returns the repo name
func (g gitlabPushEvent) GetRepoName() string {
	return g.Repo.Name
}

// GetRef returns the full ref
func (g gitlabPushEvent) GetRef() string {
	return g.Ref
}

// GetGitURL returns the git clone URL
func (g gitlabPushEvent) GetGitURL() string {
	return g.Repo.GitURL
}

// GetSSHURL returns the ssh URL
func (g gitlabPushEvent) GetSSHURL() string {
	return g.Repo.SSHURL
}