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
		return nil, errors.New("unsupported Gitlab event")
	}
}

// get payload bytes from request body
func getGitlabPayloadBytes(contentType string, body []byte) ([]byte, error) {
	switch contentType {
	case "application/json":
		return body, nil
	default:
		return nil, errors.New("Gitlab Webhook Content-Type must be application/json")
	}
}
