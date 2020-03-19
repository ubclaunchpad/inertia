package main

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/build"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/cfg"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/daemon"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
)

// Version is the current build of Inertia
var Version string

// runCmd starts the daemon
var runCmd = &cobra.Command{
	Version: getVersion(),
	Use:     "run [host]",
	Short:   "Initialize deployment and run the daemon",
	Long: `Runs the daemon on a port, default 4303. Requires
host address as an argument.

Example:
    inertia daemon run 0.0.0.0 --port 8081`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var conf = cfg.New()

		// Init webhook secret
		var webhookSecret, _ = cmd.Flags().GetString("webhook.secret")
		conf.WebhookSecret = webhookSecret

		// Set up deployment
		var projectDatabasePath = path.Join(conf.DataDirectory, "project.db")
		var projectDatabaseKeypath = path.Join(conf.SecretsDirectory, "db.key")
		deployment, err := project.NewDeployment(
			conf.ProjectDirectory,
			conf.PersistDirectory,
			projectDatabasePath,
			projectDatabaseKeypath,
			build.NewBuilder(*conf, containers.StopActiveContainers))
		if err != nil {
			println(err.Error())
			return
		}

		// Initialize daemon
		server, err := daemon.New(Version, *conf, deployment)
		if err != nil {
			println(err.Error())
			return
		}
		defer server.Close()

		var port, _ = cmd.Flags().GetString("port")
		println(server.Run(args[0], port))
	},
}

// tokenCmd retrieves the daemon token
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Produce an API token to use with the daemon",
	Long:  `Produces an API token to use with the daemon, created using an RSA private key.`,
	Run: func(cmd *cobra.Command, args []string) {
		keyBytes, err := crypto.GetAPIPrivateKey(nil)
		if err != nil {
			panic(err)
		}

		token, err := crypto.GenerateMasterToken(keyBytes.([]byte))
		if err != nil {
			panic(err)
		}

		fmt.Println(token)
	},
}

var rootCmd = &cobra.Command{
	Use:     "inertiad",
	Short:   "The inertia daemon CLI",
	Version: getVersion(),
}

func getVersion() string {
	if Version == "" {
		Version = "latest"
	}
	return Version
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(tokenCmd)
	runCmd.Flags().StringP("port", "p", "4303", "Set port for daemon to run on")
}

func main() {
	cobra.EnableCommandSorting = false
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
