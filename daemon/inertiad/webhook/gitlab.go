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

func parseGitlabEvent(r *http.Request, event string) (Payload, error) {
	dec := json.NewDecoder(r.Body)

	switch event {
	case GitlabPushHeader:
		payload := gitlabPushEvent{eventType: PushEvent}

		if err := dec.Decode(&payload); err != nil {
			return nil, errors.New("Error parsing PushEvent")
		}

		return payload, nil
	default:
		return nil, errors.New("Unsupported Gitlab event")
	}
}
