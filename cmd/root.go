package cmd

import (
	"os"

	"github.com/spf13/cobra"

	inertiacmd "github.com/ubclaunchpad/inertia/cmd/cmd"
	configcmd "github.com/ubclaunchpad/inertia/cmd/config"
	hostcmd "github.com/ubclaunchpad/inertia/cmd/host"
	"github.com/ubclaunchpad/inertia/cmd/inpututil"
	"github.com/ubclaunchpad/inertia/cmd/printutil"
	provisioncmd "github.com/ubclaunchpad/inertia/cmd/provision"
	remotecmd "github.com/ubclaunchpad/inertia/cmd/remote"

	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
)

func getVersion(version string) string {
	if version == "" {
		version = "latest"
	}
	return version
}

// NewInertiaCmd is a new Inertia command
func NewInertiaCmd(version string) *inertiacmd.Cmd {
	cobra.EnableCommandSorting = false

	// instantiate top-level command
	var root = &inertiacmd.Cmd{}
	root.Command = &cobra.Command{
		Use:     "inertia",
		Version: getVersion(version),
		Short:   "Inertia is a continuous-deployment scaffold",
		Long: `Inertia provides a continuous deployment scaffold for applications.

Initialization involves preparing a server to run an application, then
activating a daemon which will continuously update the production server
with new releases as they become available in the project's repository.

Once you have set up a remote with 'inertia remote add [remote]', use 
'inertia [remote] --help' to see what you can do with your remote. To list
available remotes, use 'inertia remote ls'.

Repository:    https://github.com/ubclaunchpad/inertia/
Issue tracker: https://github.com/ubclaunchpad/inertia/issues`,
	}

	// persistent flags across all children
	root.PersistentFlags().StringVar(&root.ConfigPath, "config", "inertia.toml", "specify relative path to Inertia configuration")
	// hack in flag parsing - this must be done because we need to initialize the
	// host commands properly when Cobra first constructs the command tree, which
	// occurs before the built-in flag parser
	for i, arg := range os.Args {
		if arg == "--config" {
			root.ConfigPath = os.Args[i+1]
			break
		}
	}

	// attach children to root 'inertia' command
	attachInitCmd(root)
	configcmd.AttachConfigCmd(root)
	remotecmd.AttachRemoteCmd(root)
	provisioncmd.AttachProvisionCmd(root)
	hostcmd.AttachHostCmds(root)

	return root
}

func attachInitCmd(inertia *inertiacmd.Cmd) {
	var init = &cobra.Command{
		Use:   "init",
		Short: "Initialize an Inertia project in this repository",
		Long: `Initializes an Inertia project in this GitHub repository.
		There must be a local git repository in order for initialization
		to succeed.`,
		Run: func(cmd *cobra.Command, args []string) {
			version := cmd.Parent().Version
			givenVersion, err := cmd.Flags().GetString("version")
			if err != nil {
				printutil.Fatal(err)
			}
			if givenVersion != version {
				version = givenVersion
			}

			// Determine best build type for project
			var buildType string
			var buildFilePath string
			cwd, err := os.Getwd()
			if err != nil {
				printutil.Fatal(err)
			}
			// docker-compose projects will usually have Dockerfiles,
			// so check for that first, then check for Dockerfile
			if common.CheckForDockerCompose(cwd) {
				println("docker-compose project detected")
				buildType = "docker-compose"
				buildFilePath = "docker-compose.yml"
			} else if common.CheckForDockerfile(cwd) {
				println("Dockerfile project detected")
				buildType = "dockerfile"
				buildFilePath = "Dockerfile"
			} else {
				println("No build file detected")
				buildType, buildFilePath, err = inpututil.AddProjectWalkthrough(os.Stdin)
				if err != nil {
					printutil.Fatal(err)
				}
			}

			// Hello world config file!
			err = local.InitializeInertiaProject(inertia.ConfigPath, inertia.Version, buildType, buildFilePath)
			if err != nil {
				printutil.Fatal(err)
			}
			println("An inertia.toml configuration file has been created to store")
			println("Inertia configuration. It is recommended that you DO NOT commit")
			println("this file in source control since it will be used to store")
			println("sensitive information.")
			println("\nYou can now use 'inertia remote add' to connect your remote")
			println("VPS instance.")
		},
	}
	init.Flags().String("version", inertia.Version, "specify Inertia daemon version to use")
	inertia.AddCommand(init)
}
