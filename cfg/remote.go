package cfg

import (
	"errors"
)

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

// Remote contains parameters for the VPS
type Remote struct {
	Version string `toml:"version"`

	IP     string  `toml:"ip"`
	SSH    *SSH    `toml:"ssh"`
	Daemon *Daemon `toml:"daemon"`

	Profiles map[string]string `toml:"profiles"`
}

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

// GetSSHHost creates the user@ip string for executing SSH commands
func (r *Remote) GetSSHHost() (string, error) {
	if r.SSH == nil {
		return "", errors.New("SSH configuration not set for remote")
	}
	return r.SSH.User + "@" + r.IP, nil
}

// GetDaemonAddr creates the IP:Port string for making requests to the Daemon
func (r *Remote) GetDaemonAddr() (string, error) {
	if r.Daemon == nil {
		return "", errors.New("Daemon configuration not set for remote")
	}
	return "https://" + r.IP + ":" + r.Daemon.Port, nil
}
