package cfg

// Inertia denotes all your configured remotes (global)
type Inertia struct {
	Remotes map[string]Remote `toml:"remotes"`
}

// Init sets up Inertia configuration
func Init(path string) (*Inertia, error) {
	return nil, nil
}

// GetRemote retrieves a remote by name
func (i *Inertia) GetRemote(name string) (*Remote, bool) {
	var remote Remote
	var ok bool
	if remote, ok = i.Remotes[name]; !ok {
		return nil, false
	}
	return &remote, true
}

// AddRemote adds a remote to configuration
func (i *Inertia) AddRemote(name string, remote Remote) bool {
	if _, ok := i.Remotes[name]; ok {
		return false
	}
	i.Remotes[name] = remote
	return true
}

// RemoveRemote removes remote with given name
func (i *Inertia) RemoveRemote(name string) bool {
	if _, ok := i.Remotes[name]; !ok {
		return false
	}
	delete(i.Remotes, name)
	return true
}
