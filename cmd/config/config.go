package configcmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/printutil"
	"github.com/ubclaunchpad/inertia/local"
)

// ConfigCmd is the parent class for the 'config' subcommands
type ConfigCmd struct {
	*cobra.Command
	projectConfigPath string
}

// AttachConfigCmd attaches the 'config' subcommands to the given parent
func AttachConfigCmd(inertia *core.Cmd) {
	var config = ConfigCmd{
		Command: &cobra.Command{
			Use:   "config [command]",
			Short: "Update and configure Inertia project settings",
			Long: `Update and configure Inertia settings pertaining to this project.

For configuring remote settings, use 'inertia remote'.`,
		},
		projectConfigPath: inertia.ProjectConfigPath,
	}
	config.attachSetCmd()
	config.attachUpgradeCmd()
	inertia.AddCommand(config.Command)
}

func (root *ConfigCmd) attachSetCmd() {
	var set = &cobra.Command{
		Use:   "set [property] [value]",
		Short: "Update a property of your Inertia project configuration",
		Long:  `Updates a property of your Inertia project configuration and save it to inertia.toml.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Ensure project initialized.
			config, err := local.GetProject(root.projectConfigPath)
			if err != nil {
				printutil.Fatal(err)
			}
			if err := cfg.SetProperty(args[0], args[1], config); err != nil {
				if err := local.Write(root.projectConfigPath, config); err != nil {
					printutil.Fatal(err)
				}
				println("Configuration setting '" + args[0] + "' has been updated.")
			} else {
				println("Configuration setting '" + args[0] + "' not found.")
			}
		},
	}
	root.AddCommand(set)
}

func (root *ConfigCmd) attachUpgradeCmd() {
	const flagVersion = "version"
	const flagRemote = "remote"
	var upgrade = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade your Inertia configuration version to match the CLI",
		Long:  `Upgrade your Inertia configuration version to match the CLI and saves it to inertia.toml`,
		Run: func(cmd *cobra.Command, args []string) {
			// Ensure project initialized.
			config, err := local.GetInertiaConfig()
			if err != nil {
				printutil.Fatal(err)
			}

			var version = root.Version
			if v, _ := cmd.Flags().GetString(flagVersion); v != "" {
				version = v
			}

			var remotes, _ = cmd.Flags().GetStringArray(flagRemote)
			if len(remotes) == 0 {
				fmt.Printf("Setting Inertia config to version '%s' for all remotes", version)
				for _, r := range config.Remotes {
					r.Version = version
				}
			} else {
				fmt.Printf("Setting Inertia config to version '%s' for remotes %s",
					version, strings.Join(remotes, ", "))
				for _, n := range remotes {
					if r, ok := config.Remotes[n]; ok {
						r.Version = version
					} else {
						printutil.Fatalf("could not find remote '%s'", n)
					}
				}
			}
			if err = local.Write(root.projectConfigPath, config); err != nil {
				printutil.Fatal(err)
			}
		},
	}
	upgrade.Flags().String(flagVersion, root.Version, "specify Inertia daemon version to set")
	upgrade.Flags().StringArrayP(flagVersion, "r", nil, "specify which remotes to modify (default: all)")
	root.AddCommand(upgrade)
}
