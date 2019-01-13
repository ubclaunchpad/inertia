package inertiacmd

import "github.com/spf13/cobra"

// Cmd is the Inertia CLI commands
type Cmd struct {
	*cobra.Command
	ConfigPath string
}
