package main

import (
	"fmt"
	"os"

	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertia/auth"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Version is the current build of Inertia
var Version string

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
		run(args[0], port, Version)
	},
}

// tokenCmd represents the daemon run command
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Produce an API token to use with the daemon",
	Long: `Produce an API token to use with the daemon,
	Created using an RSA private key.`,
	Run: func(cmd *cobra.Command, args []string) {
		keyBytes, err := auth.GetAPIPrivateKey(nil)
		if err != nil {
			log.Fatal(err)
		}

		token, err := auth.GenerateToken(keyBytes.([]byte))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(token)
	},
}

var rootCmd = &cobra.Command{
	Use:     "inertia",
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
	runCmd.Flags().StringP("port", "p", common.DefaultPort, "Set port for daemon to run on")
}

func main() {
	cobra.EnableCommandSorting = false
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
