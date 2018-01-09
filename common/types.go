package common

const (
	// DefaultSecret used for some verification
	DefaultSecret = "inertia"

	// DaemonOkResp is the OK response upon successfully reaching daemon
	DaemonOkResp = "I'm a little Webhook, short and stout!"
)

// UpRequest is the body of a up request to the daemon.
type UpRequest struct {
	Repo string `json:"repo"`
}
