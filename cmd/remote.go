package cmd

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/local"
)

// Initialize 'inertia remote [COMMAND]' commands
func init() {
	Root.AddCommand(cmdRemote)
	cmdRemote.AddCommand(cmdAddRemote)
	cmdRemote.AddCommand(cmdListRemotes)
	cmdRemote.AddCommand(cmdRemoveRemote)
	cmdRemote.AddCommand(cmdShowRemote)
	cmdRemote.AddCommand(cmdSetRemoteProperty)

	cmdListRemotes.Flags().BoolP("verbose", "v", false, "enable verbose output")
	cmdAddRemote.Flags().StringP("port", "p", "4303", "remote daemon port")
	cmdAddRemote.Flags().StringP("sshPort", "s", "22", "remote SSH port")
}

var cmdRemote = &cobra.Command{
	Use:   "remote",
	Short: "Configure the local settings for a remote host",
	Long: `Configures local settings for a remote host - add, remove, and list configured
Inertia remotes.

Requires Inertia to be set up via 'inertia init'.

For example:
    inertia init
    inertia remote add gcloud
    inertia gcloud init        # set up Inertia
	inertia gcloud status      # check on status of Inertia daemon
`,
}

var cmdAddRemote = &cobra.Command{
	Use:   "add [remote]",
	Short: "Add a reference to a remote VPS instance",
	Long: `Adds a reference to a remote VPS instance. Requires information about the VPS
including IP address, user and a PEM file. The provided name will be used in other
Inertia commands.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, path, err := local.GetProjectConfigFromDisk(configFilePath)
		if err != nil {
			log.Fatal(err)
		}

		_, found := config.GetRemote(args[0])
		if found {
			log.Fatal(errors.New("Remote " + args[0] + " already exists."))
		}

		port, _ := cmd.Flags().GetString("port")
		sshPort, _ := cmd.Flags().GetString("sshPort")
		branch, err := local.GetRepoCurrentBranch()
		if err != nil {
			log.Fatal(err)
		}

		// Start prompts and save configuration
		err = addRemoteWalkthrough(os.Stdin, config, args[0], port, sshPort, branch)
		if err != nil {
			log.Fatal(err)
		}
		err = config.Write(path)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("\nRemote '" + args[0] + "' has been added!")
		fmt.Println("You can now run 'inertia " + args[0] + " init' to set this remote up")
		fmt.Println("for continuous deployment.")
	},
}

var cmdListRemotes = &cobra.Command{
	Use:   "ls",
	Short: "List currently configured remotes",
	Long:  `Lists all currently configured remotes.`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		config, _, err := local.GetProjectConfigFromDisk(configFilePath)
		if err != nil {
			log.Fatal(err)
		}

		for _, remote := range config.Remotes {
			if verbose {
				fmt.Println(formatRemoteDetails(remote))
			} else {
				fmt.Println(remote.Name)
			}
		}
	},
}

var cmdRemoveRemote = &cobra.Command{
	Use:   "rm [remote]",
	Short: "Remove a configured remote",
	Long:  `Remove a remote from Inertia's configuration file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, path, err := local.GetProjectConfigFromDisk(configFilePath)
		if err != nil {
			log.Fatal(err)
		}

		_, found := config.GetRemote(args[0])
		if found {
			config.RemoveRemote(args[0])
			err = config.Write(path)
			if err != nil {
				log.Fatal("Failed to remove remote: " + err.Error())
			}
			fmt.Println("Remote " + args[0] + " removed.")
		} else {
			log.Fatal(errors.New("There does not appear to be a remote with this name. Have you modified the Inertia configuration file?"))
		}
	},
}

var cmdShowRemote = &cobra.Command{
	Use:   "show [remote]",
	Short: "Show details about a remote",
	Long:  `Shows details about the given remote.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, _, err := local.GetProjectConfigFromDisk(configFilePath)
		if err != nil {
			log.Fatal(err)
		}

		remote, found := config.GetRemote(args[0])
		if found {
			fmt.Println(formatRemoteDetails(remote))
		} else {
			println("No remote '" + args[0] + "' currently set up.")
		}
	},
}

var cmdSetRemoteProperty = &cobra.Command{
	Use:   "set [remote] [property] [value]",
	Short: "Update details about remote",
	Long:  `Updates the given property of the given remote's configuration.`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, path, err := local.GetProjectConfigFromDisk(configFilePath)
		if err != nil {
			log.Fatal(err)
		}

		remote, found := config.GetRemote(args[0])
		if found {
			success := setProperty(args[1], args[2], remote)
			if success {
				config.Write(path)
				println("Remote '" + args[0] + "' has been updated.")
				println(formatRemoteDetails(remote))
			} else {
				// invalid input
				println("Remote setting '" + args[1] + "' not found.")
			}
		} else {
			println("No remote '" + args[0] + "' currently set up.")
		}
	},
}
