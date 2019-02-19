package core

import (
	"github.com/spf13/cobra"
)

// Cmd is parent class for all Inertia CLI commands
type Cmd struct {
	*cobra.Command
	ProjectConfigPath string
}
