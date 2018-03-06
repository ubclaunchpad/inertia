package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is the current build of Inertia
var Version string

var rootCmd = &cobra.Command{
	Use:     "inertia",
	Short:   "Inertia is a continuous-deployment scaffold",
	Version: getVersion(),
	Long: `Inertia provides a continuous-deployment scaffold for applications.
Initialization involves preparing a server to run an application, then
activating a daemon which will continuously update the production server
with new releases as they become available in the project's repository.

One you have set up a remote with 'inertia remote add [REMOTE]',
use 'inertia [REMOTE] --help' to see what you can do with your remote.`,
}

func getVersion() string {
	if Version == "" {
		Version = "latest"
	}
	return Version
}

func main() {
	cobra.EnableCommandSorting = false
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
