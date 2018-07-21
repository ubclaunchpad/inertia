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

func parseGithubEvent(r *http.Request, event string) (Payload, error) {
	dec := json.NewDecoder(r.Body)

	switch event {
	case GithubPushHeader:
		payload := githubPushEvent{eventType: PushEvent}

		if err := dec.Decode(&payload); err != nil {
			return nil, errors.New("Error parsing PushEvent")
		}

		return payload, nil
	default:
		return nil, errors.New("Unsupported Github event")
	}
}
