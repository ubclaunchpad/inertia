package common

const (
	// DefaultSecret used for some verification
	DefaultSecret = "inertia"

	// DaemonOkResp is the OK response upon successfully reaching daemon
	DaemonOkResp = "I'm a little Webhook, short and stout!"

	// DefaultPort defines the standard daemon port
	DefaultPort = "8081"
)

// DaemonRequest is the configurable body of a request to the daemon.
type DaemonRequest struct {
	Stream     bool        `json:"stream"`
	Container  string      `json:"container,omitempty"`
	Project    string      `json:"project"`
	GitOptions *GitOptions `json:"git_options"`
}

// GitOptions represents GitHub-related deployment options
type GitOptions struct {
	RemoteURL string `json:"remote"`
	Branch    string `json:"branch"`
}
