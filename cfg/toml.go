package cfg

// InertiaRemotes represents inertia.remotes, which contains Inertia's runtime
// configuration for this project. This file should NOT be committed.
type InertiaRemotes struct {
	Version *string               `toml:"version"`
	Remotes map[string]*RemoteVPS `toml:"remotes"`
}
