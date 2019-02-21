package cmd

import (
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/input"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/output"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
	"github.com/ubclaunchpad/inertia/local/git"
)

func attachInitCmd(inertia *core.Cmd) {
	const flagGitRemote = "git.remote"
	var init = &cobra.Command{
		Use:   "init",
		Short: "Initialize an Inertia project in this repository",
		Long: `Initializes an Inertia project in this GitHub repository.
		There must be a local git repository in order for initialization
		to succeed.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Check for global inertia configuration
			if _, err := local.GetInertiaConfig(); err != nil {
				resp, err := input.Prompt("could not find global inertia configuration - would you like to initialize it?")
				if err != nil {
					output.Fatal(err)
				}
				if resp == "y" || resp == "yes" {
					if _, err := local.Init(); err != nil {
						output.Fatal(err)
					}
				} else {
					output.Fatal("global inertia configuration is required to set up Inertia")
				}
			}

			// Check for repo
			if err := git.IsRepo("."); err != nil {
				output.Fatalf("could not find git repository: %s", err.Error())
			}

			// Get host URL
			var gitRemote, _ = cmd.Flags().GetString(flagGitRemote)
			host, err := git.GetRepoRemote(gitRemote)
			if err != nil {
				output.Fatalf("could not get git remote '%s': %s", gitRemote, err.Error())
			}

			// Prompt for branch to deploy
			branch, err := git.GetRepoCurrentBranch()
			if err != nil {
				output.Fatal(err)
			}
			if resp, err := input.Promptf("Enter the branch you would like to deploy (leave blank for '%s')",
				branch); err == nil {
				branch = resp
			}

			// Determine best build type for project
			var (
				buildType     cfg.BuildType
				buildFilePath string
			)

			// docker-compose projects will usually have Dockerfiles, so check for
			// docker-compose.yml first, then check for Dockerfile
			if common.CheckForDockerCompose(".") {
				println("docker-compose project detected")
				buildType = cfg.DockerCompose
				buildFilePath = "docker-compose.yml"
			} else if common.CheckForDockerfile(".") {
				println("Dockerfile project detected")
				buildType = cfg.Dockerfile
				buildFilePath = "Dockerfile"
			} else {
				println("No build file detected")
				var err error
				buildType, buildFilePath, err = input.AddProjectWalkthrough()
				if err != nil {
					output.Fatal(err)
				}
			}

			// Hello world config file!
			if err := local.InitProject(inertia.ProjectConfigPath, host, "TODO", cfg.Profile{
				Branch: branch,
				Build: &cfg.Build{
					Type:          buildType,
					BuildFilePath: buildFilePath,
				},
			}); err != nil {
				output.Fatal(err)
			}
			println("An inertia.toml configuration file has been created to store")
			println("Inertia configuration. It is recommended that you DO NOT commit")
			println("this file in source control since it will be used to store")
			println("sensitive information.")
			println("\nYou can now use 'inertia remote add' to connect your remote")
			println("VPS instance.")
		},
	}
	init.Flags().String(flagGitRemote, "master", "git remote to use for continuous deployment")
	inertia.AddCommand(init)
}
