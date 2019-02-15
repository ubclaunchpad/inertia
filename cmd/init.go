package cmd

import (
	"os"

	"github.com/spf13/cobra"
	inertiacmd "github.com/ubclaunchpad/inertia/cmd/cmd"
	"github.com/ubclaunchpad/inertia/cmd/inpututil"
	"github.com/ubclaunchpad/inertia/cmd/printutil"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
)

func attachInitCmd(inertia *inertiacmd.Cmd) {
	const flagVersion = "version"
	var init = &cobra.Command{
		Use:   "init",
		Short: "Initialize an Inertia project in this repository",
		Long: `Initializes an Inertia project in this GitHub repository.
		There must be a local git repository in order for initialization
		to succeed.`,
		Run: func(cmd *cobra.Command, args []string) {
			var version = inertia.Version
			if givenVersion, _ := cmd.Flags().GetString(flagVersion); givenVersion != "" {
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
			err = local.InitializeInertiaProject(inertia.ConfigPath, version, buildType, buildFilePath)
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
	init.Flags().String(flagVersion, "", "specify Inertia daemon version to use")
	inertia.AddCommand(init)
}
