package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/input"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
	"github.com/ubclaunchpad/inertia/local/git"
)

var (
	errInvalidBuildType     = errors.New("invalid build type")
	errInvalidBuildFilePath = errors.New("invalid buildfile path")
)

// addProjectWalkthrough is the command line walkthrough that asks for details
// about the project the user intends to deploy.
func addProjectWalkthrough() (
	buildType cfg.BuildType, buildFilePath string, err error,
) {
	resp, err := input.NewPrompt(nil).
		PromptFromList("build type", []string{"docker-compose", "dockerfile"}).
		GetString()
	if err != nil {
		return "", "", errInvalidBuildType
	}
	buildType, err = cfg.AsBuildType(resp)
	if err != nil {
		return "", "", err
	}

	buildFilePath, err = input.NewPrompt(nil).
		Prompt(out.C("Please enter the path to your build configuration file:", out.CY)).
		GetString()
	if err != nil || buildFilePath == "" {
		return "", "", errInvalidBuildFilePath
	}
	return
}

func attachInitCmd(inertia *core.Cmd) {
	const (
		flagGitRemote = "git.remote"
		flagGlobal    = "global"
	)
	var init = &cobra.Command{
		Use:   "init",
		Short: "Initialize an Inertia project in this repository",
		Long: `Initializes an Inertia project in this GitHub repository. You can
provide an argument as the name of your project, otherwise the name of your
current directory will be used.

There must be a local git repository in order for initialization
to succeed, unless you use the '--global' flag to initialize only
the Inertia global configuration.

See https://inertia.ubclaunchpad.com/#project-configuration for more details.`,
		Example: "inertia init my_awesome_project",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			if global, _ := cmd.Flags().GetBool(flagGlobal); global {
				if _, err := local.Initialize(); err != nil {
					out.Fatal(err)
				}
				out.Printf("global Inertia configuration intialized in '%s'", local.InertiaDir())
				return
			}

			// set up coloured writer
			var highlight = out.NewColorer(out.CY)

			// Check for global inertia configuration
			if _, err := local.GetRemotes(); err != nil {
				should, err := input.NewPrompt(nil).
					Prompt(highlight.Sf(":question: Could not find global inertia configuration in %s - would you like to initialize it? (y/N)",
						local.InertiaDir())).
					GetBool()
				if err != nil {
					out.Fatal(err)
				}
				if should {
					if _, err := local.Initialize(); err != nil {
						out.Fatal(err)
					}
					out.Printf("global Inertia configuration intialized in '%s'\n", local.InertiaRemotesPath())
				} else {
					out.Fatal("aborting: global inertia configuration is required to set up Inertia")
				}
			}

			// check for local config
			if _, err := local.GetProject(inertia.ProjectConfigPath); err == nil {
				out.Fatalf("aborting: inertia configuration already exists at %s",
					inertia.ProjectConfigPath)
			}

			// Set project name
			var project string
			if len(args) == 1 {
				project = args[0]
			} else {
				cwd, _ := os.Getwd()
				project = filepath.Base(cwd)
			}
			out.Printf("initializing project '%s'\n", project)

			// Check for repo
			if err := git.IsRepo("."); err != nil {
				out.Fatalf("could not find git repository: %s", err.Error())
			}

			// Get host URL
			var gitRemote, _ = cmd.Flags().GetString(flagGitRemote)
			host, err := git.GetRepoRemote(gitRemote)
			if err != nil {
				out.Fatalf("could not get git remote '%s': %s", gitRemote, err.Error())
			}

			// Prompt for branch to deploy
			branch, err := git.GetRepoCurrentBranch()
			if err != nil {
				out.Fatal(err)
			}
			resp, err := input.NewPrompt(&input.PromptConfig{AllowEmpty: true}).
				Promptf(":evergreen_tree: %s",
					highlight.Sf(
						"Enter the branch you would like to deploy (leave blank for '%s'):",
						branch,
					)).
				GetString()
			if err == nil && resp != "" {
				branch = resp
			}

			// Determine best build type for project
			var (
				buildType     cfg.BuildType
				buildFilePath string
			)

			// docker-compose projects will usually have Dockerfiles, so check for
			// docker-compose.yml first, then check for Dockerfile
			out.Println("detecting project type...")
			if common.CheckForDockerCompose(".") {
				out.Println("docker-compose project detected :whale:")
				buildType = cfg.DockerCompose
				buildFilePath = "docker-compose.yml"
			} else if common.CheckForDockerfile(".") {
				out.Println("Dockerfile project detected :whale:")
				buildType = cfg.Dockerfile
				buildFilePath = "Dockerfile"
			} else {
				out.Println(":question: no build file detected")
				var err error
				buildType, buildFilePath, err = addProjectWalkthrough()
				if err != nil {
					out.Fatal(err)
				}
			}
			out.Println(highlight.Sf(":hammer: Profile created with %s configuration.", buildType))

			// Hello world config file!
			out.Printf("Initializing configuration file at %s...\n", inertia.ProjectConfigPath)
			if err := local.InitProject(inertia.ProjectConfigPath, project, host, cfg.Profile{
				Branch: branch,
				Build: &cfg.Build{
					Type:          buildType,
					BuildFilePath: buildFilePath,
				},
			}); err != nil {
				out.Fatal(err)
			}

			out.Println(highlight.S(":books: An inertia.toml configuration file has been created to store project settings!"))
			out.Println("You can now use 'inertia remote add' to set up your remote VPS instance.")
		},
	}
	init.Flags().String(flagGitRemote, "origin", "git remote to use for continuous deployment")
	init.Flags().BoolP(flagGlobal, "g", false, "just initialize global inertia configuration")
	inertia.AddCommand(init)
}
