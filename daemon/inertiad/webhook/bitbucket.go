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

// Due to heavy nesting, extracting keys with type assertions is preferred
func parseBitbucketEvent(r *http.Request, event string) (Payload, error) {
	dec := json.NewDecoder(r.Body)

	switch event {
	case BitbucketPushHeader:
		var rawJSON interface{}
		if err := dec.Decode(&rawJSON); err != nil {
			return nil, err
		}

		raw := rawJSON.(map[string]interface{})

		// Extract push details
		push := raw["push"].(map[string]interface{})
		changes := push["changes"].([]interface{})
		changesObj := changes[0].(map[string]interface{})
		new := changesObj["new"].(map[string]interface{})
		branchName := new["name"].(string)

		// Extract repo details
		repo := raw["repository"].(map[string]interface{})
		fullName := repo["full_name"].(string)

		payload := bitbucketPushEvent{
			eventType:  PushEvent,
			branchName: branchName,
			fullName:   fullName,
		}

		return payload, nil
	default:
		return nil, errors.New("Unsupported Bitbucket event")
	}
}
