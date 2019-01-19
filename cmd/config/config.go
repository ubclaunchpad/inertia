package configcmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cfg"
	inertiacmd "github.com/ubclaunchpad/inertia/cmd/cmd"
	"github.com/ubclaunchpad/inertia/cmd/printutil"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
)

// ConfigCmd is the parent class for the 'config' subcommands
type ConfigCmd struct {
	*cobra.Command
	cfgPath string
}

// AttachConfigCmd attaches the 'config' subcommands to the given parent
func AttachConfigCmd(inertia *inertiacmd.Cmd) {
	var config = ConfigCmd{
		Command: &cobra.Command{
			Use:   "config [command]",
			Short: "Update and configure Inertia project settings",
			Long: `Update and configure Inertia settings pertaining to this project.

For configuring remote settings, use 'inertia remote'.`,
		},
		cfgPath: inertia.ConfigPath,
	}
	config.attachSetCmd()
	config.attachUpgradeCmd()
	config.attachResetCmd()
	inertia.AddCommand(config.Command)
}

func (root *ConfigCmd) attachResetCmd() {
	var reset = &cobra.Command{
		Use:   "reset",
		Short: "Remove inertia configuration from this repository",
		Long:  `Removes Inertia configuration files pertaining to this project.`,
		Run: func(cmd *cobra.Command, args []string) {
			println("WARNING: This will remove your current Inertia configuration")
			println("and is irreversible. Continue? (y/n)")
			var response string
			_, err := fmt.Scanln(&response)
			if err != nil {
				printutil.Fatal("invalid response - aborting")
			}
			if response != "y" {
				printutil.Fatal("aborting")
			}
			path, err := common.GetFullPath(root.cfgPath)
			if err != nil {
				printutil.Fatal(err)
			}
			os.Remove(path)
			println("Inertia configuration removed.")
		},
	}
	root.AddCommand(reset)
}

func (root *ConfigCmd) attachSetCmd() {
	var set = &cobra.Command{
		Use:   "set [property] [value]",
		Short: "Update a property of your Inertia project configuration",
		Long:  `Updates a property of your Inertia project configuration and save it to inertia.toml.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Ensure project initialized.
			config, path, err := local.GetProjectConfigFromDisk(root.cfgPath)
			if err != nil {
				printutil.Fatal(err)
			}
			success := cfg.SetProperty(args[0], args[1], config)
			if success {
				config.Write(path)
				println("Configuration setting '" + args[0] + "' has been updated..")
			} else {
				println("Configuration setting '" + args[0] + "' not found.")
			}
		},
	}
	root.AddCommand(set)
}

func (root *ConfigCmd) attachUpgradeCmd() {
	const flagVersion = "version"
	var upgrade = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade your Inertia configuration version to match the CLI",
		Long:  `Upgrade your Inertia configuration version to match the CLI and saves it to inertia.toml`,
		Run: func(cmd *cobra.Command, args []string) {
			// Ensure project initialized.
			config, path, err := local.GetProjectConfigFromDisk(root.cfgPath)
			if err != nil {
				printutil.Fatal(err)
			}

			var version = root.Version
			if v, _ := cmd.Flags().GetString(flagVersion); v != "" {
				version = v
			}

			fmt.Printf("Setting Inertia config to version '%s'", version)
			config.Version = version
			if err = config.Write(path); err != nil {
				printutil.Fatal(err)
			}
		},
	}
	upgrade.Flags().String(flagVersion, root.Version, "specify Inertia daemon version to set")
	root.AddCommand(upgrade)
}
