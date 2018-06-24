package main

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
	cmdRoot.AddCommand(cmdRemote)
	cmdRemote.AddCommand(cmdAddRemote)
	cmdRemote.AddCommand(cmdListRemotes)
	cmdRemote.AddCommand(cmdRemoveRemote)
	cmdRemote.AddCommand(cmdShowRemote)
	cmdRemote.AddCommand(cmdSetRemoteProperty)

	cmdListRemotes.Flags().BoolP("verbose", "v", false, "Verbose output")
	cmdAddRemote.Flags().StringP("port", "p", "4303", "Daemon port")
	cmdAddRemote.Flags().StringP("sshPort", "s", "22", "SSH port")
}

var cmdRemote = &cobra.Command{
	Use:   "remote",
	Short: "Configure the local settings for a remote VPS instance",
	Long: `Remote is a low level command for interacting with this VPS
over SSH. Provides functionality such as adding new remotes, removing remotes,
bootstrapping the server for deployment, running install scripts such as
installing docker, starting the Inertia daemon and other low level configuration
of the VPS. Must run 'inertia init' in your repository before using.

Example:
inerta remote add gcloud
inerta gcloud init
inerta remote status gcloud`,
	Args: cobra.MinimumNArgs(1),
}

var cmdAddRemote = &cobra.Command{
	Use:   "add [REMOTE]",
	Short: "Add a reference to a remote VPS instance",
	Long: `Add a reference to a remote VPS instance. Requires 
information about the VPS including IP address, user and a PEM
file. Specify a VPS name.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, path, err := local.GetProjectConfigFromDisk(ConfigFilePath)
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
		config, _, err := local.GetProjectConfigFromDisk(ConfigFilePath)
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
	Use:   "rm [REMOTE]",
	Short: "Remove a remote.",
	Long:  `Remove a remote from Inertia's configuration file.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, path, err := local.GetProjectConfigFromDisk(ConfigFilePath)
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
	Use:   "show [REMOTE]",
	Short: "Show details about remote.",
	Long:  `Show details about the given remote.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, _, err := local.GetProjectConfigFromDisk(ConfigFilePath)
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
	Use:   "set [REMOTE] [PROPERTY] [VALUE]",
	Short: "Set details about remote.",
	Long:  `Set details about the given remote.`,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, path, err := local.GetProjectConfigFromDisk(ConfigFilePath)
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
