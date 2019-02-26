package projectcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/output"
	"github.com/ubclaunchpad/inertia/local"
	"github.com/ubclaunchpad/inertia/local/git"
)

// ProfileCmd implements the 'inertia project profile' subcommands
type ProfileCmd struct {
	*cobra.Command
	root *ProjectCmd
}

// AttachProfileCmd attaches profile subcommands to given project
func AttachProfileCmd(root *ProjectCmd) {
	var prof = &ProfileCmd{
		Command: &cobra.Command{
			Use:     "profile",
			Short:   "Manage project profile configurations",
			Long:    "Manage profile configurations for your project",
			Aliases: []string{"pf"},
		},
		root: root,
	}
	prof.attachSetCmd()
	prof.attachApplyCmd()
	prof.attachListCmd()
	prof.attachShowCmd()

	root.AddCommand(prof.Command)
}

func (p *ProfileCmd) attachSetCmd() {
	const (
		flagBranch        = "branch"
		flagBuildType     = "build.type"
		flagBuildFilePath = "build.file"
	)
	var set = &cobra.Command{
		Use:   "set [profile]",
		Short: "Configure project profiles",
		Long: `Configures project profiles - if the given profile does not exist,
a new one is created, otherwise the existing one is overwritten.

Provide profile values via the available flags.`,
		Aliases: []string{"new", "add"},
		Args:    cobra.ExactArgs(1),
		Example: "inertia project profile set my_profile --build.type dockerfile --build.file Dockerfile.dev",
		Run: func(cmd *cobra.Command, args []string) {
			var (
				err       error
				branch, _ = cmd.Flags().GetString(flagBranch)
				bTypeS, _ = cmd.Flags().GetString(flagBuildType)
				bPath, _  = cmd.Flags().GetString(flagBuildFilePath)
			)

			if branch == "" {
				branch, err = git.GetRepoCurrentBranch()
				if err != nil {
					output.Fatal(err)
				}
			}

			bType, err := cfg.AsBuildType(bTypeS)
			if err != nil {
				output.Fatal(err)
			}

			p.root.config.SetProfile(cfg.Profile{
				Name:   args[0],
				Branch: branch,
				Build: &cfg.Build{
					Type:          bType,
					BuildFilePath: bPath,
				},
			})

			if err := local.Write(p.root.projectConfigPath, p.root.config); err != nil {
				output.Fatal(err)
			}
			fmt.Printf("profile '%s' successfully updated", args[0])
		},
	}
	set.Flags().String(flagBranch, "", "branch for profile (default: current branch)")
	set.Flags().String(flagBuildType, "", "build type for profile")
	set.MarkFlagRequired(flagBuildType)
	set.Flags().String(flagBuildFilePath, "", "relative path to build config file (e.g. 'Dockerfile')")
	set.MarkFlagRequired(flagBuildFilePath)
	p.AddCommand(set)
}

func (p *ProfileCmd) attachApplyCmd() {
	var apply = &cobra.Command{
		Use:   "apply [profile] [remote]",
		Short: "Apply a project configuration profile to a remote",
		Long: `Applies a project configuration profile to an existing remote. The applied
profile will be used whenever you run 'inertia ${remote_name} up' on the target
remote.

By default, the profile called 'default' will be used.`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if _, ok := p.root.config.GetProfile(args[0]); !ok {
				output.Fatalf("profile '%s' does not exist", args[0])
			}
			cfg, err := local.GetInertiaConfig()
			if err != nil {
				output.Fatal(err)
			}
			r, ok := cfg.GetRemote(args[1])
			if !ok {
				output.Fatalf("remote '%s' does not exist", args[1])
			}
			r.ApplyProfile(p.root.config.Name, args[0])
			if err := local.SaveRemote(r); err != nil {
				output.Fatal(err)
			}
			fmt.Printf("profile '%s' successfully applied to remote '%s'", args[0], args[1])
		},
	}
	p.AddCommand(apply)
}

func (p *ProfileCmd) attachListCmd() {
	var ls = &cobra.Command{
		Use:   "ls",
		Short: "List configured project profiles",
		Long: `List configured profiles for this project. To add new ones, use
'inertia project profile set'.`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if p.root.config.Profiles == nil {
				p.root.config.Profiles = make([]*cfg.Profile, 0)
				local.Write(p.root.projectConfigPath, p.root.config)
			}
			for _, pf := range p.root.config.Profiles {
				println(pf.Name)
			}
		},
	}
	p.AddCommand(ls)
}

func (p *ProfileCmd) attachShowCmd() {
	var show = &cobra.Command{
		Use:   "show",
		Short: "Output profile configuration",
		Long: `Prints the requested profile configuration. To add new ones, use
'inertia project profile set'.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pf, ok := p.root.config.GetProfile(args[0])
			if !ok {
				output.Fatalf("profile '%s' not found", args[0])
			}
			fmt.Printf(`* Branch:              %s
* Build.Type:          %s
* Build.BuildFilePath: %s
`, pf.Branch, pf.Build.Type, pf.Build.BuildFilePath)
		},
	}
	p.AddCommand(show)
}
