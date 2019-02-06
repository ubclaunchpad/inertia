package webhook

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Constants for the generic webhook interface
const (
	// Events
	PushEvent = "push"
	PullEvent = "pull"

	// Hosts
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
func Parse(host, eventHeader string, h http.Header, body []byte) (Payload, error) {
	contentType := h.Get("content-type")

	// get payload bytes
	payloadBytes, err := getPayloadBytes(host, contentType, body)
	if err != nil {
		return nil, err
	}

	// Decode request payloadBytes to raw JSON
	var raw interface{}
	if err := json.Unmarshal(payloadBytes, &raw); err != nil {
		return nil, err
	}
	rawJSON := raw.(map[string]interface{})

	// Parse into one of supported types
	switch host {
	case GitHub:
		return parseGithubEvent(rawJSON, eventHeader)
	case GitLab:
		return parseGitlabEvent(rawJSON, eventHeader)
	case BitBucket:
		return parseBitbucketEvent(rawJSON, eventHeader)
	default:
		return nil, errors.New("Unsupported webhook received")
	}
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

// Type returns the git host and event header of given webhook request
func Type(h http.Header) (host string, eventHeader string) {
	// Parse into one of supported types
	// Try Github
	githubEventHeader := h.Get("x-github-event")
	if len(githubEventHeader) > 0 {
		host = GitHub
		eventHeader = githubEventHeader
		return
	}

	// Try Gitlab
	gitlabEventHeader := h.Get("x-gitlab-event")
	if len(gitlabEventHeader) > 0 {
		host = GitLab
		eventHeader = gitlabEventHeader
		return
	}

	// Try Bitbucket
	userAgent := h.Get("user-agent")
	if strings.Contains(userAgent, "Bitbucket") {
		host = BitBucket
		eventHeader = h.Get("x-event-key")
	}
	return
}

// get payload bytes from request body
func getPayloadBytes(host, contentType string, body []byte) ([]byte, error) {
	switch host {
	case GitHub:
		return getGithubPayloadBytes(contentType, body)
	case GitLab:
		return getGitlabPayloadBytes(contentType, body)
	case BitBucket:
		return getBitbucketPayloadBytes(contentType, body)
	default:
		return nil, errors.New("Unsupported webhook received")
	}
}
