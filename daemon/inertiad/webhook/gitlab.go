package webhook

import (
	"encoding/json"
	"errors"
	"net/http"
)

// x-gitlab-event header values
var (
	GitlabPushHeader = "Push Hook"
)

// GitlabPushEvent represents a push to a Gitlab repository
// see https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#push-events
type GitlabPushEvent struct {
	eventType string
	Ref       string                    `json:"ref"`
	Repo      GitlabPushEventRepository `json:"repository"`
}

// GitlabPushEventRepository represents the repository object in a Gitlab PushEvent
// see https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#push-events
type GitlabPushEventRepository struct {
	Name   string `json:"name"`
	GitURL string `json:"git_http_url"`
	SSHURL string `json:"git_ssh_url"`
}

func parseGitlabEvent(r *http.Request, event string) (Payload, error) {
	dec := json.NewDecoder(r.Body)

	switch event {
	case GitlabPushHeader:
		payload := GitlabPushEvent{eventType: PushEvent}

		if err := dec.Decode(&payload); err != nil {
			return nil, errors.New("Error parsing PushEvent")
		}

		return payload, nil
	default:
		return nil, errors.New("Unsupported Gitlab event")
	}
}

// GetEventType returns the event type of the webhook
func (g GitlabPushEvent) GetEventType() string {
	return g.eventType
}

// GetRepoName returns the repo name
func (g GitlabPushEvent) GetRepoName() string {
	return g.Repo.Name
}

// GetRef returns the full ref
func (g GitlabPushEvent) GetRef() string {
	return g.Ref
}

// GetGitURL returns the git clone URL
func (g GitlabPushEvent) GetGitURL() string {
	return g.Repo.GitURL
}

// GetSSHURL returns the ssh URL
func (g GitlabPushEvent) GetSSHURL() string {
	return g.Repo.SSHURL
}
