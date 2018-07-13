package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"

	"github.com/spf13/cobra"
)

var (
	// Relative paths to configuration files
	projectConfigFilePath = "inertia.toml"
	remoteConfigFilePath  = ""

	// projectName is used to generate remoteConfigFilePath
	projectName = ""

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
		switch arg {
		case "--project":
			projectName = os.Args[i+1]
		}
	}
	return nil
}

// setConfigFullPaths turns configuration paths into full paths
func setConfigFullPaths() error {
	var err error

	// Generate full path for given relative path
	projectConfigFilePath, err = common.GetFullPath(projectConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Either generate path to remotes config, or generate full path for given
	// relative path, or attempt to read config
	if projectName != "" {
		remoteConfigFilePath = local.GetRemotesConfigFilePath(projectName)
	} else {
		proj, err := cfg.ReadProjectConfig(projectConfigFilePath)
		if err != nil {
			println(err.Error())
			return nil
		}
		if proj.Project == nil {
			log.Fatal("project configuration is missing field project-name")
		}
		remoteConfigFilePath = local.GetRemotesConfigFilePath(*proj.Project)
	}

	return nil
}

func init() {
	Root.PersistentFlags().StringVar(
		&remoteConfigFilePath, "project", "",
		"Specify the name of the project Inertia should find configuration for",
	)
}
