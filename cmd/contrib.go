package cmd

import (
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
)

func getPluginPath() string {
	priority := []string{
		"INERTIA_PLUGINSPATH",
		"GOBIN",
	}
	var pluginPath string
	for _, p := range priority {
		pluginPath = os.Getenv(p)
		if pluginPath != "" {
			break
		}
	}
	// try generating GOBIN
	if pluginPath == "" {
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			pluginPath = path.Join(gopath, "bin")
		}
	}
	return pluginPath
}

func attachContribPlugins(inertia *core.Cmd) {
	var contrib = &cobra.Command{
		Use:   "contrib [tool]",
		Short: "Utilities and plugins from inertia/contrib",
		Long: `'inertia contrib' provides a shortcut for executing Inertia utilities
and plugins from inertia/contrib. These tools are installed as separate
binaries, and follow the naming convention 'inertia-{tool_name}'. Use
with care.

Install the plugins using 'go get -u github.com/ubclaunchpad/inertia/contrib/...'.

Use $INERTIA_PLUGINSPATH to configure where Inertia should look for plugins.`,
		Args:    cobra.MinimumNArgs(1),
		Example: "inertia contrib docgen",
		Hidden:  true,

		// allow flags to be passed to plugins
		DisableFlagParsing: true,
		TraverseChildren:   false,
		Run: func(cmd *cobra.Command, args []string) {
			if args[0] == "--help" {
				cmd.Help()
				return
			}
			var (
				pluginPath = getPluginPath()
				tool       = path.Join(pluginPath, "inertia-"+args[0])
				toolArgs   []string
			)
			if len(args) > 1 {
				toolArgs = args[1:]
			}

			// check if plugin is installed
			if _, err := os.Stat(tool); os.IsNotExist(err) {
				out.Fatalf("could not find plugin '%s' - please make sure it is installed",
					tool)
			}

			// execute plugin
			var plugin = exec.Command(tool, toolArgs...)
			plugin.Stdout = os.Stdout
			plugin.Stdin = os.Stdin
			if err := plugin.Run(); err != nil {
				out.Fatal(err.Error())
			}
		},
	}
	inertia.AddCommand(contrib)
}
