package cfg

import (
	"errors"
)

// Remote contains parameters for the VPS
type Remote struct {
	Version string `toml:"version"`
	Name    string `toml:"name"`

	IP     string  `toml:"ip"`
	SSH    *SSH    `toml:"ssh"`
	Daemon *Daemon `toml:"daemon"`

	Profiles map[string]string `toml:"profiles"`
}

// SSH denotes SSH options for accessing a remote
type SSH struct {
	User    string `toml:"user"`
	PEM     string `toml:"pemfile"`
	SSHPort string `toml:"ssh-port"`
}

// Daemon contains parameters for the Daemon
type Daemon struct {
	Port          string `toml:"port"`
	Token         string `toml:"token"`
	WebHookSecret string `toml:"webhook-secret"`
	VerifySSL     bool   `toml:"verify-ssl"`
}

// Identifier implements identity.Identifier
func (r *Remote) Identifier() string { return r.Name }

// GetProfile retrieves the configured profile for the named project. If no
// profile is found, `default` is returned
func (r *Remote) GetProfile(project string) string {
	if r.Profiles == nil {
		r.Profiles = make(map[string]string)
		return "default"
	}
	profile, ok := r.Profiles[project]
	if !ok {
		return "default"
	}
	return profile
}

// ApplyProfile associates the given profile name with the given project and
// saves it in Profiles
func (r *Remote) ApplyProfile(project, profile string) {
	if r.Profiles == nil {
		r.Profiles = make(map[string]string)
	}
	if profile == "" {
		profile = "default"
	}
	r.Profiles[project] = profile
}

// SSHHost creates the user@ip string for executing SSH commands
func (r *Remote) SSHHost() (string, error) {
	if r.SSH == nil {
		return "", errors.New("SSH configuration not set for remote")
	}
	return r.SSH.User + "@" + r.IP, nil
}

// DaemonAddr creates the IP:Port string for making requests to the Daemon
func (r *Remote) DaemonAddr() (string, error) {
	if r.Daemon == nil {
		return "", errors.New("Daemon configuration not set for remote")
	}
	return "https://" + r.IP + ":" + r.Daemon.Port, nil
}
