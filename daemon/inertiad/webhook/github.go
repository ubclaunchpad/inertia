package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GithubPushEvent represents a push to a Github respository
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
}

func parseGithubEvent(r *http.Request, event string) (Payload, error) {
	// TODO: Can switch on different event types here
	fmt.Printf("Type: %s", event)
	payload := GithubPushEvent{eventType: event}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return nil, errors.New("Error decoding body")
	}

	return payload, nil
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
