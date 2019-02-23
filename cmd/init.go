package cmd

import (
	"fmt"

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
	const (
		flagGitRemote = "git.remote"
		flagGlobal    = "global"
	)
	var init = &cobra.Command{
		Use:   "init",
		Short: "Initialize an Inertia project in this repository",
		Long: `Initializes an Inertia project in this GitHub repository.

There must be a local git repository in order for initialization
to succeed, unless you use the '--global' flag to initialize only
the Inertia global configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			if global, _ := cmd.Flags().GetBool(flagGlobal); global {
				if _, err := local.Init(); err != nil {
					output.Fatal(err)
				}
				fmt.Printf("global Inertia configuration intialized at %s", local.InertiaConfigPath())
				return
			}

			// Check for global inertia configuration
			if _, err := local.GetInertiaConfig(); err != nil {
				resp, err := input.Promptf("could not find global inertia configuration in %s (%s) - would you like to initialize it?",
					local.InertiaDir(), err.Error())
				if err != nil {
					output.Fatal(err)
				}
				if resp == "y" || resp == "yes" {
					if _, err := local.Init(); err != nil {
						output.Fatal(err)
					}
					fmt.Printf("global Inertia configuration intialized at %s", local.InertiaConfigPath())
				} else {
					output.Fatal("aborting: global inertia configuration is required to set up Inertia")
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
			if resp, err := input.Promptf("Enter the branch you would like to deploy (leave blank for '%s'):",
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
			if err := local.InitProject(inertia.ProjectConfigPath, "TODO", host, cfg.Profile{
				Branch: branch,
				Build: &cfg.Build{
					Type:          buildType,
					BuildFilePath: buildFilePath,
				},
			}); err != nil {
				output.Fatal(err)
			}
			println("An inertia.toml configuration file has been created to store project settings!")
			println("\nYou can now use 'inertia remote add' to set up your remote VPS instance.")
		},
	}
	init.Flags().String(flagGitRemote, "origin", "git remote to use for continuous deployment")
	init.Flags().BoolP(flagGlobal, "g", false, "just initialize global inertia configuration")
	inertia.AddCommand(init)
}
