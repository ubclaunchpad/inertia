package api

const (
	// MsgDaemonOK is the OK response upon successfully reaching daemon
	MsgDaemonOK = "I'm a little Webhook, short and stout!"

	// Container is a constant used in HTTP GET query strings
	Container = "container"

	// Stream is a constant used in HTTP GET query strings
	Stream = "stream"

	// Entries is a constant used in HTTP GET query strings
	Entries = "entries"
)

// UpRequest is the configurable body of a UP request to the daemon.
// TODO: unify with configuration definitions
type UpRequest struct {
	Stream                 bool       `json:"stream"`
	Project                string     `json:"project"`
	BuildType              string     `json:"build_type"`
	BuildFilePath          string     `json:"build_file_path"`
	GitOptions             GitOptions `json:"git_options"`
	WebHookSecret          string     `json:"webhook_secret"`
	IntermediaryContainers []string   `json:"intermediary_containers"`
	SlackNotificationURL   string     `json:"slack_notification_url"`
}

// GitOptions represents GitHub-related deployment options
type GitOptions struct {
	RemoteURL string `json:"remote"`
	Branch    string `json:"branch"`
}

// UserRequest is used for logging in or modifying users
type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Admin    bool   `json:"admin"`
	Totp     string `json:"totp"`
}

// EnvRequest represents a request to manage environment variables
type EnvRequest struct {
	Name    string `json:"name,omitempty"`
	Value   string `json:"value,omitempty"`
	Encrypt bool   `json:"encrypt,omitempty"`

	Remove bool `json:"remove,omitempty"`
}
