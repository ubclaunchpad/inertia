package webhook

// GithubPushEvent represents a push to a Github repository
// see https://developer.github.com/v3/activity/events/types/#pushevent
type githubPushEvent struct {
	eventType string
	Ref       string                    `json:"ref"`
	Repo      githubPushEventRepository `json:"repository"`
}

// GithubPushEventRepository represents the repository object in a Github PushEvent
// see https://developer.github.com/v3/activity/events/types/#pushevent
type githubPushEventRepository struct {
	Name   string `json:"name"`
	GitURL string `json:"clone_url"`
	SSHURL string `json:"ssh_url"`
}

// GetEventType returns the event type of the webhook
func (g githubPushEvent) GetEventType() string {
	return g.eventType
}

// GetRepoName returns the full repo name
func (g githubPushEvent) GetRepoName() string {
	return g.Repo.Name
}

// GetRef returns the full ref
func (g githubPushEvent) GetRef() string {
	return g.Ref
}

// GetGitURL returns the git clone URL
func (g githubPushEvent) GetGitURL() string {
	return g.Repo.GitURL
}

// GetSSHURL returns the ssh URL
func (g githubPushEvent) GetSSHURL() string {
	return g.Repo.SSHURL
}
