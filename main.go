package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is the current build of Inertia
	Version string

	// ConfigFilePath is the relative path to Inertia's configuration file
	ConfigFilePath = "inertia.toml"
)

func getVersion() string {
	if Version == "" {
		Version = "latest"
	}
	return Version
}

var cmdRoot = &cobra.Command{
	Use:     "inertia",
	Short:   "Inertia is a continuous-deployment scaffold",
	Version: getVersion(),
	Long: `Inertia provides a continuous deployment scaffold for applications.

Initialization involves preparing a server to run an application, then
activating a daemon which will continuously update the production server
with new releases as they become available in the project's repository.

One you have set up a remote with 'inertia remote add [REMOTE]',
use 'inertia [REMOTE] --help' to see what you can do with your remote.

Repository:    https://github.com/ubclaunchpad/inertia/
Issue tracker: https://github.com/ubclaunchpad/inertia/issues`,
}

func main() {
	cobra.EnableCommandSorting = false
	cmdRoot.PersistentFlags().StringVar(&ConfigFilePath, "config", "inertia.toml", "Specify relative path to Inertia configuration")
	if err := cmdRoot.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
