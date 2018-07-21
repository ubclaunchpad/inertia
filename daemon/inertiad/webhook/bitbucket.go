package webhook

import (
	"encoding/json"
	"errors"
	"net/http"
)

// x-event-key header values
var (
	BitbucketPushHeader = "repo:push"
)

func parseBitbucketEvent(r *http.Request, event string) (Payload, error) {
	dec := json.NewDecoder(r.Body)

	switch event {
	case BitbucketPushHeader:
		var raw interface{}
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}

		rawJSON := raw.(map[string]interface{})
		return parseBitbucketPushEvent(rawJSON), nil
	default:
		return nil, errors.New("Unsupported Bitbucket event")
	}
}
