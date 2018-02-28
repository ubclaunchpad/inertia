package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon"
)

// remoteCmd represents the remote command
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
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		remote, found := config.Remotes[args[0]]
		if found {
			printRemoteDetails(args[0], remote)
		}
	},
}

// addCmd represents the remote add command
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

		_, found := config.Remotes[args[0]]
		if found {
			log.Fatal(errors.New("Remote " + args[0] + " already exists."))
		}

		port, _ := cmd.Flags().GetString("port")
		sshPort, _ := cmd.Flags().GetString("sshPort")

		homeEnvVar := os.Getenv("HOME")
		sshDir := filepath.Join(homeEnvVar, ".ssh")
		defaultSSHLoc := filepath.Join(sshDir, "id_rsa")

		var response string
		fmt.Println("Enter location of PEM file (leave blank to use '" + defaultSSHLoc + "'):")
		_, err = fmt.Scanln(&response)
		if err != nil {
			response = defaultSSHLoc
		}
		pemLoc := response

		fmt.Println("Enter user:")
		_, err = fmt.Scanln(&response)
		if err != nil {
			log.Fatal("That is not a valid user - please try again.")
		}
		user := response

		fmt.Println("Enter IP address of remote:")
		_, err = fmt.Scanln(&response)
		if err != nil {
			log.Fatal("That is not a valid IP address - please try again.")
		}
		address := response

		fmt.Println("Port " + port + " will be used as the daemon port.")
		fmt.Println("Port " + sshPort + " will be used as the SSH port.")
		fmt.Println("Run 'inertia remote add' with the -p flag to set a custom Daemon port")
		fmt.Println("of the -ssh flag to set a custom SSH port.")

		err = client.AddNewRemote(args[0], address, sshPort, user, pemLoc, port)
		if err != nil {
			log.WithError(err)
		}

		fmt.Println("\nRemote '" + args[0] + "' has been added!")
		fmt.Println("You can now run 'inertia " + args[0] + " init' to set this remote up")
		fmt.Println("for continuous deployment.")
	},
}

// deployInitCmd represents the inertia [REMOTE] init command
var deployInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the VPS for continuous deployment",
	Long: `Initialize the VPS for continuous deployment.
This sets up everything you might need and brings the Inertia daemon
online on your remote.
A URL will be provided to direct GitHub webhooks to, the daemon will
request access to the repository via a public key, and will listen
for updates to this repository's remote master branch.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		remote, found := config.Remotes[remoteName]
		if found {
			session := client.NewSSHRunner(remote)
			err = remote.Bootstrap(session, "", config)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(errors.New("There does not appear to be a remote with this name. Have you modified the Inertia configuration file?"))
		}
	},
}

// statusCmd represents the remote status command
var statusCmd = &cobra.Command{
	Use:   "status [REMOTE]",
	Short: "Query the status of a remote instance",
	Long: `Query the remote VPS for connectivity, daemon
behaviour, and other information.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.WithError(err)
		}

		remote, found := config.Remotes[args[0]]
		if !found {
			println("No such remote " + args[0])
			println("Inertia currently supports one remote per repository")
			println("Run `inertia remote -v' to see what remote is available")
			os.Exit(1)
		}

		host := "http://" + remote.GetIPAndPort()
		resp, err := http.Get(host)
		if err != nil {
			println("Could not connect to daemon")
			println("Try running inertia [REMOTE] init")
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			println("Bad response from daemon")
			println("Try running inertia [REMOTE] init")
			return
		}

		if string(body) != common.DaemonOkResp {
			println("Could not connect to daemon")
			println("Try running inertia [REMOTE] init")
			return
		}

		fmt.Printf("Remote instance '%s' accepting requests at %s\n",
			args[0], host)
	},
}

// listCmd represents the inertia list command
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

		for name, remote := range config.Remotes {
			if verbose {
				printRemoteDetails(name, remote)
			} else {
				fmt.Println(name)
			}
		}
	},
}

// removeCmd represents the inertia list command
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

		_, found := config.Remotes[args[0]]
		if found {
			config.Remotes[args[0]] = nil
			delete(config.Remotes, args[0])
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

func printRemoteDetails(name string, remote *client.RemoteVPS) {
	fmt.Printf("Remote %s: \n", name)
	fmt.Printf(" - IP Address:  %s\n", remote.IP)
	fmt.Printf(" - Daemon Port: %s\n", remote.Daemon.Port)
	fmt.Printf(" - VPS User:    %s\n", remote.User)
	fmt.Printf(" - PEM File:    %s\n", remote.PEM)
}

func init() {
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(addCmd)
	remoteCmd.AddCommand(statusCmd)
	remoteCmd.AddCommand(listCmd)
	remoteCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	listCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	addCmd.Flags().StringP("port", "p", daemon.DefaultPort, "Daemon port")
	addCmd.Flags().StringP("sshPort", "s", "22", "SSH port")
}
