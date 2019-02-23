package cfg

import (
	"github.com/ubclaunchpad/inertia/cfg/internal/identity"
)

// Project represents the current project's configuration.
type Project struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`

	Profiles []*Profile `toml:"profile"`
}

// Profile denotes a deployment configuration
type Profile struct {
	Name   string `toml:"name"`
	Branch string `toml:"branch"`
	Build  *Build `toml:"build"`
}

// Identifier implements identity.Identifier
func (p *Profile) Identifier() string { return p.Name }

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
	BuildFilePath string    `toml:"buildfile"`
}

// NewProject sets up Inertia configuration with given properties
func NewProject(name, host string) *Project {
	if name == "" {
		name = "inertia-deployment"
	}
	return &Project{
		Name:     name,
		URL:      host,
		Profiles: make([]*Profile, 0),
	}
}

// GetProfile retrieves the named profile
func (p *Project) GetProfile(name string) (*Profile, bool) {
	if name == "" {
		return nil, false
	}
	v, ok := identity.Get(name, ident(p.Profiles))
	if !ok {
		return nil, false
	}
	return v.(*Profile), ok
}

// SetProfile assigns a profile to project configuration
func (p *Project) SetProfile(profile Profile) {
	if profile.Name == "" {
		return
	}
	if profile.Build == nil {
		profile.Build = &Build{}
	}
	var ids = ident(p.Profiles)
	identity.Set(&profile, &ids)
	p.Profiles = asProfiles(ids)
}

// RemoveProfile removes a configured profile
func (p *Project) RemoveProfile(name string) bool {
	var ids = ident(p.Profiles)
	ok := identity.Remove(name, &ids)
	p.Profiles = asProfiles(ids)
	return ok
}
