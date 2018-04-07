package common

const (
	// DaemonOkResp is the OK response upon successfully reaching daemon
	DaemonOkResp = "I'm a little Webhook, short and stout!"
)

// DaemonRequest is the configurable body of a request to the daemon.
type DaemonRequest struct {
	Stream     bool        `json:"stream"`
	Container  string      `json:"container,omitempty"`
	Project    string      `json:"project"`
	BuildType  string      `json:"build_type"`
	GitOptions *GitOptions `json:"git_options"`
	Secret     string      `json:"secret"`
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
}
