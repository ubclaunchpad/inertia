package projectcmd

import (
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/output"
	"github.com/ubclaunchpad/inertia/local"
)

// ProjectCmd is the parent class for the 'config' subcommands
type ProjectCmd struct {
	*cobra.Command
	projectConfigPath string
}

// AttachProjectCmd attaches the 'config' subcommands to the given parent
func AttachProjectCmd(inertia *core.Cmd) {
	var project = &ProjectCmd{
		Command: &cobra.Command{
			Use:   "project [command]",
			Short: "Update and configure Inertia project settings",
			Long: `Update and configure Inertia settings pertaining to this project.

For configuring remote settings, use 'inertia remote'.`,
		},
		projectConfigPath: inertia.ProjectConfigPath,
	}
	project.attachSetCmd()
	AttachProfileCmd(project)

	inertia.AddCommand(project.Command)
}

func (root *ProjectCmd) attachSetCmd() {
	var set = &cobra.Command{
		Use:   "set [property] [value]",
		Short: "Update a property of your Inertia project configuration",
		Long:  `Updates a property of your Inertia project configuration and save it to inertia.toml.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Ensure project initialized.
			config, err := local.GetProject(root.projectConfigPath)
			if err != nil {
				output.Fatal(err)
			}
			if err := cfg.SetProperty(args[0], args[1], config); err != nil {
				if err := local.Write(root.projectConfigPath, config); err != nil {
					output.Fatal(err)
				}
				println("Configuration setting '" + args[0] + "' has been updated.")
			} else {
				println("Configuration setting '" + args[0] + "' not found.")
			}
		},
	}
	root.AddCommand(set)
}
