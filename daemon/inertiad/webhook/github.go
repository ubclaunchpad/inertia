package webhook

import (
	"errors"
)

// x-github-event header values
var (
	GithubPushHeader = "push"
	// GithubPullHeader = "pull"
)

func parseGithubEvent(rawJSON map[string]interface{}, event string) (Payload, error) {
	switch event {
	case GithubPushHeader:
		return parseGithubPushEvent(rawJSON), nil
	default:
		return nil, errors.New("Unsupported Github event")
	}
}
