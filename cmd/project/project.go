package projectcmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/input"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/output"
	"github.com/ubclaunchpad/inertia/local"
)

// ProjectCmd is the parent class for the 'config' subcommands
type ProjectCmd struct {
	*cobra.Command
	config            *cfg.Project
	projectConfigPath string
}

// AttachProjectCmd attaches the 'config' subcommands to the given parent
func AttachProjectCmd(inertia *core.Cmd) {
	var project = &ProjectCmd{
		projectConfigPath: inertia.ProjectConfigPath,
	}
	project.Command = &cobra.Command{
		Use:   "project [command]",
		Short: "Update and configure Inertia project settings",
		Long: `Update and configure Inertia settings pertaining to this project.

To create a new project, use 'inertia init'.

For configuring remote settings, use 'inertia remote'.`,
		PersistentPreRun: func(*cobra.Command, []string) {
			var err error
			project.config, err = local.GetProject(project.projectConfigPath)
			if err != nil {
				fmt.Printf("could not find project configuration at '%s': %s",
					project.projectConfigPath, err.Error())
				output.Fatal("try instantiating a new project using 'inertia init'")
			}
		},
	}
	AttachProfileCmd(project)
	project.attachSetCmd()
	project.attachResetCmd()

	inertia.AddCommand(project.Command)
}

func (root *ProjectCmd) attachSetCmd() {
	var set = &cobra.Command{
		Use:   "set [property] [value]",
		Short: "Update a property of your Inertia project configuration",
		Long:  `Updates a property of your Inertia project configuration and save it to inertia.toml.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cfg.SetProperty(args[0], args[1], root.config); err != nil {
				if err := local.Write(root.projectConfigPath, root.config); err != nil {
					output.Fatal(err)
				}
				println("configuration setting '" + args[0] + "' has been updated")
			} else {
				println("configuration setting '" + args[0] + "' not found")
			}
		},
	}
	root.AddCommand(set)
}

func (root *ProjectCmd) attachResetCmd() {
	var reset = &cobra.Command{
		Use:   "reset",
		Short: "Remove project configuration",
		Long: `Removes your project configuration by deleting the configuration file.
	This is irreversible.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := input.Prompt("would you like to reset your project configuration?")
			if err != nil {
				output.Fatal(err)
			}
			if resp == "y" || resp == "yes" {
				if err := os.Remove(root.projectConfigPath); err != nil {
					output.Fatal(err)
				}
			} else {
				output.Fatal("aborting")
			}
		},
	}
	root.AddCommand(reset)
}
