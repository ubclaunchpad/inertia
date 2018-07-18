package common

const (
	// MsgDaemonOK is the OK response upon successfully reaching daemon
	MsgDaemonOK = "I'm a little Webhook, short and stout!"

	// Container is a constant used in HTTP GET query strings
	Container = "container"

	// Stream is a constant used in HTTP GET query strings
	Stream = "stream"
)

// UpRequest is the configurable body of a UP request to the daemon.
type UpRequest struct {
	Stream        bool        `json:"stream"`
	WebHookSecret string      `json:"webhook_secret"`
	GitOptions    *GitOptions `json:"git_options"`
}

// GitOptions represents GitHub-related deployment options
type GitOptions struct {
	Branch string `json:"branch"`
}

// UserRequest is used for logging in or modifying users
type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Admin    bool   `json:"admin"`
}

// EnvRequest represents a request to manage environment variables
type EnvRequest struct {
	Name    string `json:"name,omitempty"`
	Value   string `json:"value,omitempty"`
	Encrypt bool   `json:"encrypt,omitempty"`

	Remove bool `json:"remove,omitempty"`
}
