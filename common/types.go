package common

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
