package webhook

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The endpoint does not really matter, we are only interested in
// how the request body gets parsed by the Webhook package
func getMockRequest(endpoint string, contentType string, rawBody []byte) *http.Request {
	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(rawBody))
	req.Header.Add("Content-Type", contentType)
	return req
}

func TestTypeAndParse(t *testing.T) {
	testCases := []struct {
		source      string
		contentType string
		reqBody     []byte
		eventHeader string
		eventValue  string
	}{
		{GitHub, "application/x-www-form-urlencoded", githubPushFormEncoded, "x-github-event", GithubPushHeader},
		{GitHub, "application/json", githubPushRawJSON, "x-github-event", GithubPushHeader},
		{GitLab, "application/json", gitlabPushRawJSON, "x-gitlab-event", GitlabPushHeader},
		{BitBucket, "application/json", bitbucketPushRawJSON, "x-event-key", BitbucketPushHeader},
	}
	for _, tc := range testCases {
		req := getMockRequest("/webhook", tc.contentType, tc.reqBody)
		req.Header.Add(tc.eventHeader, tc.eventValue)

		// Special case for Bitbucket because Bitbucket
		if tc.eventHeader == "x-event-key" {
			req.Header.Add("User-Agent", "Bitbucket")
		}

		// Parse type
		host, event := Type(req.Header)

		// Read
		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err)

		// Parse payload
		payload, err := Parse(host, event, req.Header, body)
		assert.Nil(t, err)

		assert.Equal(t, tc.source, payload.GetSource())
		assert.Equal(t, "push", payload.GetEventType())
		assert.Equal(t, "inertia-deploy-test", payload.GetRepoName())
		assert.Equal(t, "refs/heads/master", payload.GetRef())
	}
}

func TestParseDocker(t *testing.T) {
	req := getMockRequest("/docker-webhook", "application/json", dockerPushRawJSON)
	payload, err := ParseDocker(req)
	assert.Nil(t, err)

	assert.Equal(t, "briannguyen", payload.GetPusher())
	assert.Equal(t, "latest", payload.GetTag())
	assert.Equal(t, "ubclaunchpad/inertia", payload.GetRepoName())
	assert.Equal(t, "inertia", payload.GetName())
	assert.Equal(t, "ubclaunchpad", payload.GetOwner())
}
