package cfg

// TOML formats are defined here, used by Config.Write()

// InertiaProject represents inertia.toml, which contains Inertia's persistent
// project configurations. This file should be committed.
type InertiaProject struct {
	Version       *string `toml:"version"`
	Project       *string `toml:"project-name"`
	BuildType     *string `toml:"build-type"`
	BuildFilePath *string `toml:"build-file-path"`
}

// InertiaRemotes represents inertia.remotes, which contains Inertia's runtime
// configuration for this project. This file should NOT be committed.
type InertiaRemotes struct {
	Version *string                `toml:"version"`
	Project *string                `toml:"project-name"`
	Remotes *map[string]*RemoteVPS `toml:"remotes"`
}
