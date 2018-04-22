package main

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/common"
)

var (
	errInvalidUser    = errors.New("invalid user")
	errInvalidAddress = errors.New("invalid IP address")
	errInvalidSecret  = errors.New("invalid secret")
)

var remoteCmd = &cobra.Command{
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

var addCmd = &cobra.Command{
	Use:   "add [REMOTE]",
	Short: "Add a reference to a remote VPS instance",
	Long: `Add a reference to a remote VPS instance. Requires 
information about the VPS including IP address, user and a PEM
file. Specify a VPS name.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		_, found := config.GetRemote(args[0])
		if found {
			log.Fatal(errors.New("Remote " + args[0] + " already exists."))
		}

		port, _ := cmd.Flags().GetString("port")
		sshPort, _ := cmd.Flags().GetString("sshPort")

		repo, err := common.GetLocalRepo()
		if err != nil {
			log.Fatal(err)
		}
		head, err := repo.Head()
		if err != nil {
			log.Fatal(err)
		}
		branch := head.Name().Short()

		err = addRemoteWalkthrough(os.Stdin, args[0], port, sshPort, branch, config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("\nRemote '" + args[0] + "' has been added!")
		fmt.Println("You can now run 'inertia " + args[0] + " init' to set this remote up")
		fmt.Println("for continuous deployment.")
	},
}

var listCmd = &cobra.Command{
	Use:   "ls",
	Short: "List currently configured remotes",
	Long:  `Lists all currently configured remotes.`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		for _, remote := range config.Remotes {
			if verbose {
				printRemoteDetails(remote)
			} else {
				fmt.Println(remote.Name)
			}
		}
	},
}

var removeCmd = &cobra.Command{
	Use:   "rm [REMOTE]",
	Short: "Remove a remote.",
	Long:  `Remove a remote from Inertia's configuration file.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		_, found := config.GetRemote(args[0])
		if found {
			config.RemoveRemote(args[0])
			err = config.Write()
			if err != nil {
				log.Fatal("Failed to remove remote: " + err.Error())
			}
			fmt.Println("Remote " + args[0] + " removed.")
		} else {
			log.Fatal(errors.New("There does not appear to be a remote with this name. Have you modified the Inertia configuration file?"))
		}
	},
}

var showCmd = &cobra.Command{
	Use:   "show [REMOTE]",
	Short: "Show details about remote.",
	Long:  `Show details about the given remote.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		remote, found := config.GetRemote(args[0])
		if found {
			printRemoteDetails(remote)
		} else {
			println("No remote '" + args[0] + "' currently set up.")
		}
	},
}

func printRemoteDetails(remote *client.RemoteVPS) {
	fmt.Printf("Remote %s: \n", remote.Name)
	fmt.Printf(" - Deployed Branch:   %s\n", remote.Branch)
	fmt.Printf(" - IP Address:        %s\n", remote.IP)
	fmt.Printf(" - VPS User:          %s\n", remote.User)
	fmt.Printf(" - PEM File Location: %s\n", remote.PEM)
	fmt.Printf("Run 'inertia %s status' for more details.\n", remote.Name)
}

func init() {
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(addCmd)
	remoteCmd.AddCommand(listCmd)
	remoteCmd.AddCommand(removeCmd)
	remoteCmd.AddCommand(showCmd)

	listCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	addCmd.Flags().StringP("port", "p", "8081", "Daemon port")
	addCmd.Flags().StringP("sshPort", "s", "22", "SSH port")
}
