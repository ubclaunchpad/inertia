package api

// BaseResponse is the underlying response structure to all responses.
type BaseResponse struct {
	// Basic metadata
	HTTPStatusCode int    `json:"code"`
	RequestID      string `json:"request_id,omitempty"`

	// Message is included in all responses, and is a summary of the server's response
	Message string `json:"message"`

	// Error contains additional context in the event of an error
	Error string `json:"error,omitempty"`

	// Data contains information the server wants to return
	// To parse data into a particular type, you can
	Data map[string]interface{} `json:"data,omitempty"`
}

// TotpResponse is used for sending users their Totp secret and backup codes
type TotpResponse struct {
	TotpSecret  string   `json:"secret"`
	BackupCodes []string `json:"backup_codes"`
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
