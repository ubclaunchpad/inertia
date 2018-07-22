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
		return nil, errors.New("Unsupported Bitbucket event")
	}
}
