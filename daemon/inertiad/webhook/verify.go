package webhook

import (
	"errors"
	"io/ioutil"
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
func Verify(host, key string, r *http.Request) (err error) {
	switch host {
	case BitBucket:
		// Bitbucket server has HMAC verification (same as GitHub), but not
		// the standard Bitbucket, it seems.
		if r.Header.Get(xHubSignatureHeader) == "" {
			return nil // assume the event is valid
		}
		fallthrough // use same validation as GitHub
	case GitHub:
		// https://developer.github.com/webhooks/securing/
		var payloadBytes []byte
		if payloadBytes, err = ioutil.ReadAll(r.Body); err != nil {
			return nil
		}
		return crypto.ValidateSignature(
			r.Header.Get(xHubSignatureHeader),
			payloadBytes,
			[]byte(key))
	case GitLab:
		// https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#secret-token
		token := r.Header.Get(gitlabTokenHeader)
		if token != key {
			return errors.New("invalid webhook token")
		}
		return nil
	default:
		return errors.New("unsupported type")
	}
}
