package cfg

import (
	"github.com/ubclaunchpad/inertia/cfg/internal/identity"
)

// Inertia denotes all your configured remotes (global)
type Inertia struct {
	Remotes []*Remote `toml:"remote"`
}

// NewInertiaConfig instantiates a new Inertia configuration
func NewInertiaConfig() *Inertia {
	return &Inertia{
		Remotes: make([]*Remote, 0),
	}
}

// GetRemote retrieves a remote by name
func (i *Inertia) GetRemote(name string) (*Remote, bool) {
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
func (i *Inertia) SetRemote(remote Remote) {
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
func (i *Inertia) RemoveRemote(name string) bool {
	if name == "" {
		return false
	}
	var ids = ident(i.Remotes)
	ok := identity.Remove(name, &ids)
	i.Remotes = asRemotes(ids)
	return ok
}
