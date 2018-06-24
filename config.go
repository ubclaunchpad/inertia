package main

// initCmd represents the init command
import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
)

// Initialize "inertia" commands regarding basic configuration
func init() {
	cmdRoot.AddCommand(cmdInit)
	cmdRoot.AddCommand(cmdReset)
	cmdRoot.AddCommand(cmdSetConfigProperty)

	cmdInit.Flags().String("version", Version, "Specify Inertia daemon version to use")
}

var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize an inertia project in this repository",
	Long: `Initialize an inertia project in this GitHub repository.
There must be a local git repository in order for initialization
to succeed.`,
	Run: func(cmd *cobra.Command, args []string) {
		version := cmd.Parent().Version
		givenVersion, err := cmd.Flags().GetString("version")
		if err != nil {
			log.Fatal(err)
		}
		if givenVersion != version {
			version = givenVersion
		}

		// Determine best build type for project
		var buildType string
		var buildFilePath string
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
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
		} else if common.CheckForProcfile(cwd) {
			println("Heroku project detected")
			buildType = "herokuish"
		} else {
			println("No build file detected")
			buildType, buildFilePath, err = addProjectWalkthrough(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Hello world config file!
		err = local.InitializeInertiaProject(ConfigFilePath, version, buildType, buildFilePath)
		if err != nil {
			log.Fatal(err)
		}
		println("An inertia.toml configuration file has been created to store")
		println("Inertia configuration. It is recommended that you DO NOT commit")
		println("this file in source control since it will be used to store")
		println("sensitive information.")
		println("\nYou can now use 'inertia remote add' to connect your remote")
		println("VPS instance.")
	},
}

var cmdReset = &cobra.Command{
	Use:   "reset",
	Short: "Reset the Inertia project in this repository.",
	Long: `Reset removes the existing Inertia configuration from
	this repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		println("WARNING: This will remove your current Inertia configuration")
		println("and is irreversible. Continue? (y/n)")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			log.Fatal("invalid response - aborting")
		}
		if response != "y" {
			log.Fatal("aborting")
		}
		path, err := common.GetFullPath(ConfigFilePath)
		if err != nil {
			log.Fatal(err)
		}
		os.Remove(path)
		println("Inertia configuration removed.")
	},
}

var cmdSetConfigProperty = &cobra.Command{
	Use:   "set [PROPERTY] [VALUE]",
	Short: "Set configuration property of the project",
	Long:  `Set configuration property of the project. This will modify local toml file.`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, path, err := local.GetProjectConfigFromDisk(ConfigFilePath)
		if err != nil {
			log.Fatal(err)
		}
		success := setProperty(args[0], args[1], config)
		if success {
			config.Write(path)
			println("Configuration setting '" + args[0] + "' has been updated..")
		} else {
			println("Configuration setting '" + args[0] + "' not found.")
		}
	},
}
