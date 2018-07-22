package webhook

import (
	"errors"
)

// x-gitlab-event header values
var (
	GitlabPushHeader = "Push Hook"
)

func parseGitlabEvent(rawJSON map[string]interface{}, event string) (Payload, error) {
	switch event {
	case GitlabPushHeader:
		return parseGitlabPushEvent(rawJSON), nil
	default:
		return nil, errors.New("Unsupported Gitlab event")
	}
}
