package cfg

import (
	"github.com/ubclaunchpad/inertia/cfg/internal/identity"
)

// Remotes denotes global Inertia configuration
type Remotes struct {
	// Remotes tracks globally configured remotes. It is a list instead of a map
	// to better align with TOML best practices
	Remotes []*Remote `toml:"remote"`
}

// NewRemotesConfig instantiates a new Inertia configuration
func NewRemotesConfig() *Remotes {
	return &Remotes{
		Remotes: make([]*Remote, 0),
	}
}

// GetRemote retrieves a remote by name
func (i *Remotes) GetRemote(name string) (*Remote, bool) {
	if name == "" {
		return nil, false
	}
	v, ok := identity.Get(name, ident(i.Remotes))
	if !ok {
		return nil, false
	}
	return v.(*Remote), ok
}

// SetRemote adds or updates a remote to configuration
func (i *Remotes) SetRemote(remote Remote) {
	if remote.Name == "" {
		return
	}
	if remote.Daemon == nil {
		remote.Daemon = &Daemon{}
	}
	var ids = ident(i.Remotes)
	identity.Set(&remote, &ids)
	i.Remotes = asRemotes(ids)
}

// RemoveRemote removes remote with given name
func (i *Remotes) RemoveRemote(name string) bool {
	if name == "" {
		return false
	}
	var ids = ident(i.Remotes)
	ok := identity.Remove(name, &ids)
	i.Remotes = asRemotes(ids)
	return ok
}
