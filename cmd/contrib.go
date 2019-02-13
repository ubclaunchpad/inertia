package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	inertiacmd "github.com/ubclaunchpad/inertia/cmd/cmd"
	"github.com/ubclaunchpad/inertia/cmd/printutil"
)

func attachContribPlugins(inertia *inertiacmd.Cmd) {
	var contrib = &cobra.Command{
		Use:   "contrib [tool]",
		Short: "Utilities and plugins from inertia/contrib",
		Long: `'inertia contrib' provides a shortcut for executing Inertia
utilities and plugins. These tools are installed as separate binaries,
and follow the naming convention 'inertia-{tool_name}'. Use these with
care.

Install the plugins using 'go get -u github.com/ubclaunchpad/inertia/contrib/...'.

Use $INERTIA_PLUGINSPATH to configure where Inertia should look for plugins.`,
		Args:    cobra.MinimumNArgs(1),
		Example: "inertia contrib docgen",
		Hidden:  true,

		// allow flags to be passed to plugins
		DisableFlagParsing: true,
		TraverseChildren:   false,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				path     = os.Getenv("INERTIA_PLUGINSPATH")
				tool     = path + "inertia-" + args[0]
				toolArgs []string
			)
			if len(args) > 1 {
				toolArgs = args[1:]
			}

			// check if plugin is installed
			if _, err := os.Stat(tool); os.IsNotExist(err) {
				printutil.Fatalf("could not find plugin '%s' - please make sure it is installed",
					tool)
			}

			// execute plugin
			var plugin = exec.Command(tool, toolArgs...)
			plugin.Stdout = os.Stdout
			plugin.Stdin = os.Stdin
			if err := plugin.Run(); err != nil {
				printutil.Fatal(err.Error())
			}
		},
	}
	inertia.AddCommand(contrib)
}
