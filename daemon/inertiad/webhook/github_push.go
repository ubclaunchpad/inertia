package webhook

// Implements Payload interface
// See github_test.go for an example request body
type githubPushEvent struct {
	eventType EventType
	ref       string
	name      string
	gitURL    string
	sshURL    string
}

func parseGithubPushEvent(rawJSON map[string]interface{}) githubPushEvent {
	// Extract push details
	// First level contains ref and repo
	ref := rawJSON["ref"].(string)
	repo := rawJSON["repository"].(map[string]interface{})

	// Extract repo details
	name := repo["name"].(string)
	gitURL := repo["clone_url"].(string)
	sshURL := repo["ssh_url"].(string)

	return githubPushEvent{
		eventType: PushEvent,
		ref:       ref,
		name:      name,
		gitURL:    gitURL,
		sshURL:    sshURL,
	}
}

// GetSource returns the source of the webhook
func (g githubPushEvent) GetSource() string {
	return GitHub
}

// GetEventType returns the event type of the webhook
func (g githubPushEvent) GetEventType() EventType {
	return g.eventType
}

// GetRepoName returns the full repo name
func (g githubPushEvent) GetRepoName() string {
	return g.name
}

// GetRef returns the full ref
func (g githubPushEvent) GetRef() string {
	return g.ref
}

// GetGitURL returns the git clone URL
func (g githubPushEvent) GetGitURL() string {
	return g.gitURL
}

// GetSSHURL returns the ssh URL
func (g githubPushEvent) GetSSHURL() string {
	return g.sshURL
}
