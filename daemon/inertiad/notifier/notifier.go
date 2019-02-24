package notifier

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Notifier manages notifications
type Notifier interface {
	Notify(string) error
}

// SlackNotifier represents slack notifications
type SlackNotifier struct {
	hookURL string
}

// NewNotifier creates a notifier with web hook url to slack channel
func NewNotifier() *SlackNotifier {
	url := "https://hooks.slack.com/services/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"

	n := &SlackNotifier{
		hookURL: url,
	}

	return n
}

// Notify sends the notification
func (n *SlackNotifier) Notify(text string) error {
	message := map[string]interface{}{
		"text": text,
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		return err
	}

	resp, err := http.Post(n.hookURL, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return err
	}

	_ = resp //note: temporary, may need response in the future?
	/*
		var result map[string]interface{}

		json.NewDecoder(resp.Body).Decode(&result)

		return result["data"].(string), nil
	*/
	return err
}
