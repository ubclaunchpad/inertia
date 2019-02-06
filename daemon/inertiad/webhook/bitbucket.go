package webhook

import (
	"errors"
)

// x-event-key header values
var (
	BitbucketPushHeader = "repo:push"
)

func parseBitbucketEvent(rawJSON map[string]interface{}, event string) (Payload, error) {
	switch event {
	case BitbucketPushHeader:
		return parseBitbucketPushEvent(rawJSON), nil
	default:
		return nil, errors.New("unsupported Bitbucket event")
	}
}

// get payload bytes from request body
func getBitbucketPayloadBytes(contentType string, body []byte) ([]byte, error) {
	switch contentType {
	case "application/json":
		return body, nil
	default:
		return nil, errors.New("Bitbucket Webhook Content-Type must be application/json")
	}
}
