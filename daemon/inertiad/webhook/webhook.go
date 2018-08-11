package webhook

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Constants for the generic webhook interface
var (
	PushEvent = "push"
	// PullEvent = "pull"

	GitHub    = "github"
	GitLab    = "gitlab"
	BitBucket = "bitbucket"
)

// Payload represents a generic webhook payload
type Payload interface {
	GetSource() string
	GetEventType() string
	GetRepoName() string
	GetRef() string
	GetGitURL() string
	GetSSHURL() string
}

// Parse takes in a webhook request and parses it into one of the supported types
func Parse(r *http.Request) (Payload, error) {
	// Decode request body to raw JSON
	if r.Header.Get("content-type") != "application/json" {
		return nil, errors.New("Webhook Content-Type must be JSON")
	}

	var raw interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		return nil, err
	}
	rawJSON := raw.(map[string]interface{})

	// Parse into one of supported types
	// Try Github
	githubEventHeader := r.Header.Get("x-github-event")
	if len(githubEventHeader) > 0 {
		return parseGithubEvent(rawJSON, githubEventHeader)
	}

	// Try Gitlab
	gitlabEventHeader := r.Header.Get("x-gitlab-event")
	if len(gitlabEventHeader) > 0 {
		return parseGitlabEvent(rawJSON, gitlabEventHeader)
	}

	// Try Bitbucket
	userAgent := r.Header.Get("user-agent")
	if strings.Contains(userAgent, "Bitbucket") {
		bitbucketEventHeader := r.Header.Get("x-event-key")
		return parseBitbucketEvent(rawJSON, bitbucketEventHeader)
	}

	return nil, errors.New("Unsupported webhook received")
}

// ParseDocker takes in a Docker webhook request and parses it
func ParseDocker(r *http.Request) (*DockerWebhook, error) {
	// Decode request body to raw JSON
	if r.Header.Get("content-type") != "application/json" {
		return nil, errors.New("Webhook Content-Type must be JSON")
	}

	var raw interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		return nil, err
	}
	rawJSON := raw.(map[string]interface{})

	return parseDocker(rawJSON)
}
