package notify

import (
	"go.uber.org/multierr"
)

// Notifiers is a collection of notification targets
type Notifiers []Notifier

// Notify delivers a notification to all targets
func (n Notifiers) Notify(msg string, opts Options) error {
	if len(n) == 0 {
		return nil
	}

	var errs error
	for _, notif := range n {
		errs = multierr.Append(errs, notif.Notify(msg, opts))
	}
	return errs
}

// Exists checks if the given notifier is already configured
func (n Notifiers) Exists(nt Notifier) bool {
	for _, notif := range n {
		if notif.IsEqual(nt) {
			return true
		}
	}
	return false
}

// Notifier manages notifications
type Notifier interface {
	Notify(string, Options) error
	IsEqual(Notifier) bool
}

// Options is used to configure formatting of notifications
type Options struct {
	Color Color
}

// Color is used to represent message color for different states (i.e success, fail)
type Color string

const (
	// Green for success messages
	Green Color = "good"
	// Yellow for warning messages
	Yellow Color = "warning"
	// Red for error messages
	Red Color = "danger"
)

// MessageArray builds the json message to be posted to webhook
type MessageArray struct {
	Attachments []Message `json:"attachments"`
}

// Message builds the attachments content of Message
type Message struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}
