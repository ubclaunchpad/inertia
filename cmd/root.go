package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/ubclaunchpad/inertia/common"

	"github.com/spf13/cobra"
)

var (
	// Relative paths to configuration files
	projectConfigFilePath = "inertia.toml"
	remoteConfigFilePath  = "inertia.remotes"

	// Setup - these functions run before init()
	parse       = parseConfigArgs()
	setFullPath = setConfigFullPaths()
)

// Root is the base inertia command
var Root = &cobra.Command{
	Use:   "inertia",
	Short: "Inertia is a continuous-deployment scaffold",
	Long: `Inertia provides a continuous deployment scaffold for applications.

Initialization involves preparing a server to run an application, then
activating a daemon which will continuously update the production server
with new releases as they become available in the project's repository.

One you have set up a remote with 'inertia remote add [REMOTE]',
use 'inertia [REMOTE] --help' to see what you can do with your remote.

Repository:    https://github.com/ubclaunchpad/inertia/
Issue tracker: https://github.com/ubclaunchpad/inertia/issues`,
	Version: "latest",
}

// parseConfigArgs is a dirty dirty hack to allow access to the --config argument
// before Cobra parses it (it is required to set up remote commands in the
// init() phase)
func parseConfigArgs() error {
	for i, arg := range os.Args {
		if arg == "--project" {
			projectConfigFilePath = os.Args[i+1]
		}
		if arg == "--remotes" {
			remoteConfigFilePath = os.Args[i+1]
		}
	}
	return nil
}

// setConfigFullPaths turns configuration paths into full paths
func setConfigFullPaths() error {
	var err error
	projectConfigFilePath, err = common.GetFullPath(projectConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}
	remoteConfigFilePath, err = common.GetFullPath(remoteConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func init() {
	Root.PersistentFlags().StringVar(
		&projectConfigFilePath, "project", "inertia.toml",
		"Specify relative path to Inertia project configuration",
	)
	Root.PersistentFlags().StringVar(
		&remoteConfigFilePath, "remotes", "inertia.remotes",
		"Specify relative path to Inertia remotes configuration",
	)
}
