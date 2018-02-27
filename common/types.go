package common

const (
	// DefaultSecret used for some verification
	DefaultSecret = "inertia"

	// DaemonOkResp is the OK response upon successfully reaching daemon
	DaemonOkResp = "I'm a little Webhook, short and stout!"
)

// DaemonRequest is the configurable body of a request to the daemon.
type DaemonRequest struct {
	Stream    bool   `json:"stream"`
	Repo      string `json:"repo,omitempty"`
	Container string `json:"container,omitempty"`
}
