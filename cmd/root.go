package cmd

import "github.com/spf13/cobra"

var (
	// configFilePath is the relative path to Inertia's configuration file
	configFilePath = "inertia.toml"

	// Version is the current build of Inertia
	Version string
)

func getVersion() string {
	if Version == "" {
		Version = "latest"
	}
	return Version
}

// Root is the base inertia command
var Root = &cobra.Command{
	Use:   "inertia",
	Short: "Inertia is a continuous-deployment scaffold",
	Long: `Inertia provides a continuous deployment scaffold for applications.

Initialization involves preparing a server to run an application, then
activating a daemon which will continuously update the production server
with new releases as they become available in the project's repository.

Once you have set up a remote with 'inertia remote add [remote]', use 
'inertia [remote] --help' to see what you can do with your remote.

Repository:    https://github.com/ubclaunchpad/inertia/
Issue tracker: https://github.com/ubclaunchpad/inertia/issues`,
	Version: getVersion(),
}
