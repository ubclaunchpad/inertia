package webhook

import (
	"errors"
	"net/http"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
)

const (
	// Signatures
	xHubSignatureHeader = "X-Hub-Signature"
	gitlabTokenHeader   = "X-Gitlab-Token"
)

// Verify ensures the payload's integrity and returns and error if anything
// doesn't match up
func Verify(host, key string, h http.Header, body []byte) (err error) {
	switch host {
	case BitBucket:
		// Bitbucket server has HMAC verification (same as GitHub), but not
		// the standard Bitbucket, it seems.
		if h.Get(xHubSignatureHeader) == "" {
			return nil // assume the event is valid
		}
		fallthrough // use same validation as GitHub
	case GitHub:
		// https://developer.github.com/webhooks/securing/
		return crypto.ValidateSignature(
			h.Get(xHubSignatureHeader),
			body,
			[]byte(key))
	case GitLab:
		// https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#secret-token
		token := h.Get(gitlabTokenHeader)
		if token != key {
			return errors.New("invalid webhook token")
		}
		return nil
	default:
		return errors.New("unsupported type")
	}
}
