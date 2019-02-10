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

// BaseResponse is the underlying response structure to all responses
type BaseResponse struct {
	// Basic metadata
	HTTPStatusCode int    `json:"code"`
	RequestID      string `json:"request_id,omitempty"`

	// Message is included in all responses, and is a summary of the server's response
	Message string `json:"message"`

	// Error contains additional context in the event of an error
	Error string `json:"error,omitempty"`

	// Data contains information the server wants to return
	Data map[string]interface{} `json:"data,omitempty"`
}

// UpRequest is the configurable body of a UP request to the daemon.
type UpRequest struct {
	Stream        bool       `json:"stream"`
	Project       string     `json:"project"`
	BuildType     string     `json:"build_type"`
	BuildFilePath string     `json:"build_file_path"`
	GitOptions    GitOptions `json:"git_options"`
	WebHookSecret string     `json:"webhook_secret"`
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

// TotpResponse is used for sending users their Totp secret and backup codes
type TotpResponse struct {
	TotpSecret  string   `json:"secret"`
	BackupCodes []string `json:"backup_codes"`
}

// EnvRequest represents a request to manage environment variables
type EnvRequest struct {
	Name    string `json:"name,omitempty"`
	Value   string `json:"value,omitempty"`
	Encrypt bool   `json:"encrypt,omitempty"`

	Remove bool `json:"remove,omitempty"`
}

// DeploymentStatus lists details about the deployed project
type DeploymentStatus struct {
	InertiaVersion       string   `json:"version"`
	Branch               string   `json:"branch"`
	CommitHash           string   `json:"commit_hash"`
	CommitMessage        string   `json:"commit_message"`
	BuildType            string   `json:"build_type"`
	Containers           []string `json:"containers"`
	BuildContainerActive bool     `json:"build_active"`
}
