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
		case "--config":
			projectConfigFilePath = os.Args[i+1]
		case "--project":
			projectName = os.Args[i+1]
		case "--remotes":
			remoteConfigFilePath = os.Args[i+1]
		}
	}
	if projectName != "" && remoteConfigFilePath != "" {
		log.Fatal("cannot set both --project and --remotes")
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
	} else if remoteConfigFilePath != "" {
		remoteConfigFilePath, err = common.GetFullPath(remoteConfigFilePath)
		if err != nil {
			log.Fatal(err)
		}
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
		&projectConfigFilePath, "config", "inertia.toml",
		"Specify relative path to Inertia project configuration",
	)
	Root.PersistentFlags().StringVar(
		&remoteConfigFilePath, "remotes", "",
		"Specify relative path to remote configuration file",
	)
	Root.PersistentFlags().StringVar(
		&remoteConfigFilePath, "project", "",
		"Specify the name of the project Inertia should find configuration for",
	)
}
