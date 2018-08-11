package webhook

import (
	"fmt"
	"strings"
)

// Implements Payload interface
// See bitbucket_test.go for an example request body
type bitbucketPushEvent struct {
	eventType  string
	branchName string
	fullName   string
}

func parseBitbucketPushEvent(rawJSON map[string]interface{}) bitbucketPushEvent {
	// Extract push details - branch name is retrieved
	push := rawJSON["push"].(map[string]interface{})
	changes := push["changes"].([]interface{})
	changesObj := changes[0].(map[string]interface{})
	new := changesObj["new"].(map[string]interface{})
	branchName := new["name"].(string)

	// Extract repo details -- full name is retrieved
	repo := rawJSON["repository"].(map[string]interface{})
	fullName := repo["full_name"].(string)
	return bitbucketPushEvent{
		eventType:  PushEvent,
		branchName: branchName,
		fullName:   fullName,
	}
}

// GetSource returns the source of the webhook
func (b bitbucketPushEvent) GetSource() string {
	return BitBucket
}

// GetEventType returns the event type of the webhook
func (b bitbucketPushEvent) GetEventType() string {
	return b.eventType
}

// GetRepoName returns the full repo name
// full name takes the form [user]/[repo]
func (b bitbucketPushEvent) GetRepoName() string {
	return strings.Split(b.fullName, "/")[1]
}

// GetRef returns the full ref
func (b bitbucketPushEvent) GetRef() string {
	return fmt.Sprintf("refs/heads/%s", b.branchName)
}

// GetGitURL returns the git clone URL
// Ex. https://ubclaunchpad@bitbucket.org/ubclaunchpad/inertia.git
func (b bitbucketPushEvent) GetGitURL() string {
	user := strings.Split(b.fullName, "/")[0]
	return fmt.Sprintf("https://%s@bitbucket.org/%s.git", user, b.fullName)
}

// GetSSHURL returns the ssh URL
// Ex. git@bitbucket.org:ubclaunchpad/inertia.git
func (b bitbucketPushEvent) GetSSHURL() string {
	return "git@bitbucket.org:" + b.fullName + ".git"
}
