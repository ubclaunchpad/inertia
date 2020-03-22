package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// SlackNotifier represents slack notifications
type SlackNotifier struct {
	hookURL string
}

// NewSlackNotifier creates a notifier with web hook url to slack channel. Passing
// it an empty url makes it a no-op notifier.
func NewSlackNotifier(webhookURL string) Notifier {
	return &SlackNotifier{
		hookURL: webhookURL,
	}
}

// Notify sends the notification
func (n *SlackNotifier) Notify(text string, options Options) error {
	if n.hookURL == "" {
		return nil
	}

	b, err := json.Marshal(MessageArray{
		Attachments: []Message{
			{
				Text:  fmt.Sprintf("*%s*", text),
				Color: colorToString(options.Color),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	resp, err := http.Post(n.hookURL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New("http request rejected by Slack: " + string(body))
	}

	return nil
}

// IsEqual implements Notifier by checking the provided notifier is a slack notifier
// and if it has the same hook URL
func (n *SlackNotifier) IsEqual(nt Notifier) bool {
	switch v := nt.(type) {
	case *SlackNotifier:
		return n.hookURL == v.hookURL
	default:
		return false
	}
}

func colorToString(color Color) string {
	return string(color)
}
