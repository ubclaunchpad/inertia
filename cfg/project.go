package cfg

import "fmt"

// Project represents the current project's configuration.
type Project struct {
	Version string `toml:"version"`
	Project string `toml:"project"`

	Profiles map[string]Profile `toml:"profiles"`
}

// Profile denotes a deployment configuration
type Profile struct {
	Branch string `toml:"branch"`
	Build  *Build `toml:"build"`
}

// BuildType represents supported build types
type BuildType string

const (
	// Dockerfile is used for plain Dockerfile builds
	Dockerfile BuildType = "dockerfile"

	// DockerCompose is used for docker-compose configurations
	DockerCompose BuildType = "docker-compose"
)

// Build denotes build configuration
type Build struct {
	Type          BuildType `toml:"type"`
	BuildFilePath string    `toml:"build_file"`
}

// NewProject sets up Inertia configuration with given properties
func NewProject(version, project string) *Project {
	if version == "" {
		version = "latest"
	}
	if project == "" {
		project = "inertia-deployment"
	}
	return &Project{
		Version:  version,
		Project:  project,
		Profiles: make(map[string]Profile),
	}
}

// SetProfile assigns a profile to project configuration
func (p *Project) SetProfile(name string, profile Profile) {
	if profile.Build == nil {
		profile.Build = &Build{}
	}
	p.Profiles[name] = profile
}

// RemoveProfile removes a configured profile
func (p *Project) RemoveProfile(name string) error {
	if _, ok := p.Profiles[name]; !ok {
		return fmt.Errorf("could not find profile '%s'", name)
	}
	delete(p.Profiles, name)
	return nil
}
