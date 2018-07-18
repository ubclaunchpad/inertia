package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"

	"github.com/spf13/cobra"
)

var (
	// projectName is used to generate remoteConfigFilePath
	projectName string

	// Setup - these functions run before init()
	parse = parseConfigArgs()

	// Relative paths to configuration files
	projectConfigFilePath, remoteConfigFilePath = setConfigFullPaths()
)

// Root is the base inertia command
var Root = &cobra.Command{
	Use:   "inertia",
	Short: "Inertia is a continuous-deployment scaffold",
	Long: `Inertia provides a continuous deployment scaffold for applications.

Initialization involves preparing a server to run an application, then
activating a daemon which will continuously update the production server
with new releases as they become available in the project's repository.

One you have set up a remote with 'inertia remote add [remote]', use 
'inertia [remote] --help' to see what you can do with your remote.

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
func setConfigFullPaths() (projectConfigFilePath string, remoteConfigFilePath string) {
	var err error

	// Generate full path for given relative path
	projectConfigFilePath, err = common.GetFullPath("inertia.toml")
	if err != nil {
		log.Fatal(err)
	}

	// Either generate path to remotes config, or generate full path for given
	// relative path, or attempt to read config
	if projectName != "" {
		remoteConfigFilePath = local.GetRemotesConfigFilePath(projectName)
	} else {
		proj, err := common.ReadProjectConfig(projectConfigFilePath)
		if err != nil {
			println(err.Error())
			return
		}
		if proj.Project == nil {
			log.Fatal("project configuration is missing field 'project-name'")
		}
		projectName = *proj.Project
		remoteConfigFilePath = local.GetRemotesConfigFilePath(projectName)
	}

	return
}

func init() {
	Root.PersistentFlags().StringVar(
		&remoteConfigFilePath, "project", "",
		"Specify the name of the project Inertia should find configuration for",
	)
}
