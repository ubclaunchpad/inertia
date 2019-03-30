package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
func NewNotifier(webhook string) *SlackNotifier {
	url := webhook

	if webhook == "test" { // temporary for testing
		url = "https://hooks.slack.com/services/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"
	}
	// os.Getenv("SLACK_URL")

	n := &SlackNotifier{
		hookURL: url,
	}

	return n
}

// Color is used to represent message color for different states (i.e success, fail)
type Color string

const (
	// Green when build successful
	Green Color = "good"
	// Yellow ...
	Yellow Color = "warning"
	// Red when build unsuccessful
	Red Color = "danger"
)

// NotifyOptions is used to configure formatting of notifications
type NotifyOptions struct {
	Color   Color
	Warning bool
}

// MessageArray builds the json message to be posted to webhook
type MessageArray struct {
	Attachments []Message `json:"attachments"`
}

// Message builds the attachments content of Message
type Message struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

// Notify sends the notification
func (n *SlackNotifier) Notify(text string, options *NotifyOptions) error {
	// check if url is empty
	if n.hookURL == "" {
		return nil
	}

	msg := MessageArray{
		Attachments: []Message{
			{
				Text:  "*" + text + "*" + "\nCheck details <blank_url>",
				Color: colorToString(options.Color),
			},
		},
	}
	bytesRepresentation, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to encode request: %s", err.Error())
	}

	resp, err := http.Post(n.hookURL, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	bodyString := string(body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New("Http request rejected by Slack. Error: " + bodyString)
	}

	return nil
}

func colorToString(color Color) string {
	return string(color)
}
