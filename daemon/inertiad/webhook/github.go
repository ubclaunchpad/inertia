package webhook

import (
	"errors"
	"fmt"
	"net/url"
)

// x-github-event header values
var (
	GithubPingHeader = "ping"
	GithubPushHeader = "push"
	// GithubPullHeader = "pull"
)

func parseGithubEvent(rawJSON map[string]interface{}, event string) (Payload, error) {
	switch event {
	case GithubPingHeader:
		return githubPushEvent{eventType: PingEvent}, nil
	case GithubPushHeader:
		return parseGithubPushEvent(rawJSON), nil
	default:
		return nil, fmt.Errorf("unsupported Github event %s", event)
	}
}

// get payload bytes from request body
func getGithubPayloadBytes(contentType string, body []byte) ([]byte, error) {
	switch contentType {
	case "application/x-www-form-urlencoded":
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
		return []byte(values.Get("payload")), nil
	case "application/json":
		return body, nil
	default:
		return nil, errors.New("Github Webhook Content-Type must be application/json or application/x-www-form-urlencoded")
	}
}
