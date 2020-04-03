package projectcmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/input"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
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
				out.Printf("could not find project configuration at '%s': %s",
					project.projectConfigPath, err.Error())
				out.Fatal("try instantiating a new project using 'inertia init'")
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
			if err := cfg.SetProperty(args[0], args[1], root.config); err == nil {
				if err := local.Write(root.projectConfigPath, root.config); err != nil {
					out.Fatal(err)
				}
				out.Println("configuration setting '" + args[0] + "' has been updated")
			} else {
				out.Println("configuration setting '" + args[0] + "' could not be updated: " + err.Error())
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
		Run: func(cmd *cobra.Command, args []string) {
			should, err := input.NewPrompt(nil).
				Prompt("Would you like to reset your project configuration? (y/N)").
				GetBool()
			if err != nil {
				out.Fatal(err)
			}
			if should {
				if err := os.Remove(root.projectConfigPath); err != nil {
					out.Fatal(err)
				}
			} else {
				out.Fatal("aborting")
			}
		},
	}
	root.AddCommand(reset)
}
