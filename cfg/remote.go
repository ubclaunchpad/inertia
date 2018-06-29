package cfg

// RemoteVPS contains parameters for the VPS
type RemoteVPS struct {
	Name    string        `toml:"name"`
	IP      string        `toml:"IP"`
	User    string        `toml:"user"`
	PEM     string        `toml:"pemfile"`
	Branch  string        `toml:"branch"`
	SSHPort string        `toml:"ssh-port"`
	Daemon  *DaemonConfig `toml:"daemon"`
}

// DaemonConfig contains parameters for the Daemon
type DaemonConfig struct {
	Port          string `toml:"port"`
	Token         string `toml:"token"`
	WebHookSecret string `toml:"webhook-secret"`
}

// GetHost creates the user@IP string.
func (remote *RemoteVPS) GetHost() string {
	return remote.User + "@" + remote.IP
}

// GetIPAndPort creates the IP:Port string.
func (remote *RemoteVPS) GetIPAndPort() string {
	return remote.IP + ":" + remote.Daemon.Port
}
