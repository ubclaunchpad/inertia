package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Constants for the generic webhook interface
var (
	PushEvent = "push"
	// PullEvent = "pull"
)

// Payload represents a generic webhook payload
type Payload interface {
	GetEventType() string
	GetRepoName() string
	GetRef() string
	GetGitURL() string
	GetSSHURL() string
}

// Parse takes in a webhook request and parses it into one of the supported types
func Parse(r *http.Request, out io.Writer) (Payload, error) {
	if r.Header.Get("content-type") != "application/json" {
		return nil, errors.New("Webhook Content-Type must be JSON")
	}

	var raw interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		return nil, err
	}
	rawJSON := raw.(map[string]interface{})

	// Try Github
	githubEventHeader := r.Header.Get("x-github-event")
	if len(githubEventHeader) > 0 {
		fmt.Fprintln(out, "Github webhook detected")
		return parseGithubEvent(rawJSON, githubEventHeader)
	}

	// Try Gitlab
	gitlabEventHeader := r.Header.Get("x-gitlab-event")
	if len(gitlabEventHeader) > 0 {
		fmt.Fprintln(out, "Gitlab webhook detected")
		return parseGitlabEvent(rawJSON, gitlabEventHeader)
	}

	// Try Bitbucket
	userAgent := r.Header.Get("user-agent")
	if strings.Contains(userAgent, "Bitbucket") {
		fmt.Fprintln(out, "Bitbucket webhook detected")
		bitbucketEventHeader := r.Header.Get("x-event-key")
		return parseBitbucketEvent(rawJSON, bitbucketEventHeader)
	}

	return nil, errors.New("Unsupported webhook received")
}

// ParseDocker takes in a Docker webhook request and parses it
func ParseDocker(r *http.Request, out io.Writer) (DockerWebhook, error) {
	var payload DockerWebhook
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		fmt.Println(err.Error())
		return payload, errors.New("Unable to parse Docker Webhook")
	}

	return payload, nil
}
