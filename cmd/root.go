package cmd

import (
	"os"

	"github.com/spf13/cobra"

	configcmd "github.com/ubclaunchpad/inertia/cmd/config"
	"github.com/ubclaunchpad/inertia/cmd/core"
	hostcmd "github.com/ubclaunchpad/inertia/cmd/host"
	provisioncmd "github.com/ubclaunchpad/inertia/cmd/provision"
	remotecmd "github.com/ubclaunchpad/inertia/cmd/remote"
)

func getVersion(version string) string {
	if version == "" {
		version = "latest"
	}
	return version
}

// NewInertiaCmd is a new Inertia command
func NewInertiaCmd(version string) *core.Cmd {
	cobra.EnableCommandSorting = false

	// instantiate top-level command
	var root = &core.Cmd{}
	root.Command = &cobra.Command{
		Use:     "inertia",
		Version: getVersion(version),
		Short:   "Effortless, self-hosted continuous deployment for small teams and projects",
		Long: `Inertia is an effortless, self-hosted continuous deployment platform.

Initialization involves preparing a server to run an application, then
activating a daemon which will continuously update the production server
with new releases as they become available in the project's repository.

Once you have set up a remote with 'inertia remote add [remote]', use 
'inertia [remote] --help' to see what you can do with your remote. To list
available remotes, use 'inertia remote ls'.

Repository:    https://github.com/ubclaunchpad/inertia/
Issue tracker: https://github.com/ubclaunchpad/inertia/issues`,
		DisableAutoGenTag: true,
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
	attachContribPlugins(root)

	return root
}
