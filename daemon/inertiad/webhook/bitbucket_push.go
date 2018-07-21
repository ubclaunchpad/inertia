package webhook

import "strings"

// bitbucketPushEvent represents a push to a Bitbucket repository
// see https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html
type bitbucketPushEvent struct {
	eventType  string
	branchName string
	fullName   string
}

// Due to heavy nesting, extracting keys with type assertions is preferred
func parseBitbucketPushEvent(rawJSON map[string]interface{}) bitbucketPushEvent {
	// Extract push details
	push := rawJSON["push"].(map[string]interface{})
	changes := push["changes"].([]interface{})
	changesObj := changes[0].(map[string]interface{})
	new := changesObj["new"].(map[string]interface{})
	branchName := new["name"].(string)

	// Extract repo details
	repo := rawJSON["repository"].(map[string]interface{})
	fullName := repo["full_name"].(string)
	return bitbucketPushEvent{
		eventType:  PushEvent,
		branchName: branchName,
		fullName:   fullName,
	}
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
	return "refs/heads/" + b.branchName
}

// GetGitURL returns the git clone URL
// Ex. https://ubclaunchpad@bitbucket.org/ubclaunchpad/inertia.git
func (b bitbucketPushEvent) GetGitURL() string {
	user := strings.Split(b.fullName, "/")[0]
	return "https://" + user + "@bitbucket.org/" + b.fullName + ".git"
}

// GetSSHURL returns the ssh URL
// Ex. git@bitbucket.org:ubclaunchpad/inertia.git
func (b bitbucketPushEvent) GetSSHURL() string {
	return "git@bitbucket.org:" + b.fullName + ".git"
}
