package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

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

inerta remote add gcloud 35.123.55.12 -i /Users/path/to/pem
inerta remote bootstrap gcloud
inerta remote status gcloud`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}
		if config.CurrentRemoteName == client.NoInertiaRemote {
			println("No remote currently set.")
		} else if verbose {
			fmt.Printf("%s\n", config.CurrentRemoteName)
			fmt.Printf("%+v\n", config.CurrentRemoteVPS)
		} else {
			println(config.CurrentRemoteName)
		}
	},
}

// addCmd represents the remote add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a reference to a remote VPS instance",
	Long: `Add a reference to a remote VPS instance. Requires 
information about the VPS including IP address, user and a PEM
file. Specify a VPS name and an IP address.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		_, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.WithError(err)
		}
		user, _ := cmd.Flags().GetString("user")
		pemLoc, _ := cmd.Flags().GetString("identity")
		port, _ := cmd.Flags().GetString("port")
		err = client.AddNewRemote(args[0], args[1], user, pemLoc, port)
		if err != nil {
			log.WithError(err)
		}
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
		// TODO: support correct remote based on which
		// cmd is calling this init, see "deploy.go"

		// Ensure project initialized.
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		session := client.NewSSHRunner(config.CurrentRemoteVPS)
		err = config.CurrentRemoteVPS.Bootstrap(session, "", config)
		if err != nil {
			log.Fatal(err)
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

		if args[0] != config.CurrentRemoteName {
			println("No such remote " + args[0])
			println("Inertia currently supports one remote per repository")
			println("Run `inertia remote -v' to see what remote is available")
			os.Exit(1)
		}

		host := "http://" + config.CurrentRemoteVPS.GetIPAndPort()
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
			config.CurrentRemoteName, host)
	},
}

func init() {
	RootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(addCmd)
	remoteCmd.AddCommand(statusCmd)

	homeEnvVar := os.Getenv("HOME")
	sshDir := filepath.Join(homeEnvVar, ".ssh")
	defaultSSHLoc := filepath.Join(sshDir, "id_rsa")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	remoteCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	addCmd.Flags().StringP("user", "u", "root", "User for SSH access")
	addCmd.Flags().StringP("identity", "i", defaultSSHLoc, "PEM file location")
	addCmd.Flags().StringP("port", "p", daemon.DefaultPort, "Daemon port")
}
