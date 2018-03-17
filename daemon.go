package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/daemon"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Configure daemon behaviour from command line",
	Long: `Configure daemon behaviour from the command line.
This is intended for use on a remote VPS - do not use these commands
locally.`,
	Args: cobra.MinimumNArgs(1),
	Run:  func(cmd *cobra.Command, args []string) {},
}

// runCmd represents the daemon run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the daemon",
	Long: `Run the daemon on a port.
Example:

inertia daemon run -p 8081`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetString("port")
		if err != nil {
			log.WithError(err)
		}
		daemon.Run(args[0], port, Version)
	},
}

// tokenCmd represents the daemon run command
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Produce an API token to use with the daemon",
	Long: `Produce an API token to use with the daemon,
	Created using an RSA private key.`,
	Run: func(cmd *cobra.Command, args []string) {
		keyBytes, err := daemon.GetAPIPrivateKey(nil)
		if err != nil {
			log.Fatal(err)
		}

		token, err := daemon.GenerateToken(keyBytes.([]byte))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(token)
	},
}

func init() {
	if os.Getenv("INERTIA_DAEMON") == "true" {
		rootCmd.AddCommand(daemonCmd)
		daemonCmd.AddCommand(runCmd)
		daemonCmd.AddCommand(tokenCmd)
		// Here you will define your flags and configuration settings.

		// Cobra supports Persistent Flags which will work for this command
		// and all subcommands, e.g.:
		// daemonCmd.PersistentFlags().String("foo", "", "A help for foo")

		// Cobra supports local flags which will only run when this command
		// is called directly, e.g.:
		runCmd.Flags().StringP("port", "p", "8081", "Set port for daemon to run on")
	}
}
