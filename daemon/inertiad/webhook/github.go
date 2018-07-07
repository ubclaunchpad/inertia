package webhook

import (
	"encoding/json"
	"errors"
	"net/http"
)

// x-github-event header values
var (
	GithubPushHeader = "push"
	// GithubPullHeader = "pull"
)

// GithubPushEvent represents a push to a Github repository
// see https://developer.github.com/v3/activity/events/types/#pushevent
type GithubPushEvent struct {
	eventType string
	Ref       string                    `json:"ref"`
	Repo      GithubPushEventRepository `json:"repository"`
}

// GithubPushEventRepository represents the repository object in a Github PushEvent
// see https://developer.github.com/v3/activity/events/types/#pushevent
type GithubPushEventRepository struct {
	FullName string `json:"full_name"`
	GitURL   string `json:"clone_url"`
	SSHURL   string `json:"ssh_url"`
}

func parseGithubEvent(r *http.Request, event string) (Payload, error) {
	dec := json.NewDecoder(r.Body)

	switch event {
	case GithubPushHeader:
		payload := GithubPushEvent{eventType: PushEvent}

		if err := dec.Decode(&payload); err != nil {
			return nil, errors.New("Error parsing PushEvent")
		}

		return payload, nil
	default:
		return nil, errors.New("Unsupported Github event")
	}
}

// GetEventType returns the event type of the webhook
func (g GithubPushEvent) GetEventType() string {
	return g.eventType
}

// GetRepoName returns the full repo name
func (g GithubPushEvent) GetRepoName() string {
	return g.Repo.FullName
}

// GetRef returns the full ref
func (g GithubPushEvent) GetRef() string {
	return g.Ref
}

// GetGitURL returns the git clone URL
func (g GithubPushEvent) GetGitURL() string {
	return g.Repo.GitURL
}

// GetSSHURL returns the ssh URL
func (g GithubPushEvent) GetSSHURL() string {
	return g.Repo.SSHURL
}
