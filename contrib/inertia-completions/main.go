package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ubclaunchpad/inertia/cmd"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
	"github.com/ubclaunchpad/inertia/local"
)

// Version denotes the version of the binary
var Version string

func main() {
	os.Setenv(out.EnvColorToggle, "false")
	var root = cmd.NewInertiaCmd(Version, local.InertiaDir(), false)
	if err := newDocgenCmd(root).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func newDocgenCmd(root *core.Cmd) *cobra.Command {
	const (
		flagFormat = "format"
	)
	var docs = &cobra.Command{
		Use:     "inertia-completions [dir]",
		Hidden:  true,
		Version: Version,
		Args:    cobra.MinimumNArgs(1),
		Example: "inertia contrib completions ${fpath[1]} -f zsh",
		Short:   "Generate completions for the Inertia CLI.",
		Long:    `Generate completions for the Inertia CLI. Supports bash and zsh.`,
		Run: func(cmd *cobra.Command, args []string) {
			outPath := args[0]
			var format, _ = cmd.Flags().GetString(flagFormat)

			switch format {
			case "zsh":
				// https://github.com/spf13/cobra/blob/master/zsh_completions.md
				if err := root.GenZshCompletionFile(outPath); err != nil {
					out.Fatal(err)
				}
			case "bash":
				// https://github.com/spf13/cobra/blob/master/bash_completions.md
				if err := root.GenBashCompletionFile(outPath); err != nil {
					out.Fatal(err)
				}
			default:
				out.Fatalf("unsupported completions format %s (allowed: bash, zsh)", format)
			}

			out.Printf("%s completions generated in %s\n", format, outPath)
		},
	}
	docs.Flags().StringP(flagFormat, "f", "bash", "format to generate (bash|zsh)")
	return docs
}
