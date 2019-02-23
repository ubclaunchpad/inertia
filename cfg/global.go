package cfg

import (
	"github.com/ubclaunchpad/inertia/cfg/internal/identity"
)

// Inertia denotes all your configured remotes (global)
type Inertia struct {
	Remotes []*Remote `toml:"remotes"`
}

// NewInertiaConfig instantiates a new Inertia configuration
func NewInertiaConfig() *Inertia {
	return &Inertia{
		Remotes: make([]*Remote, 0),
	}
}

// GetRemote retrieves a remote by name
func (i *Inertia) GetRemote(name string) (*Remote, bool) {
	r, ok := identity.Find(name, ident(i.Remotes))
	if !ok {
		return nil, false
	}
	return r.(*Remote), ok
}

// SetRemote adds or updates a remote to configuration
func (i *Inertia) SetRemote(remote Remote) {
	identity.Set(&remote, ident(i.Remotes))
}

// RemoveRemote removes remote with given name
func (i *Inertia) RemoveRemote(name string) bool {
	return identity.Remove(name, ident(i.Remotes))
}
