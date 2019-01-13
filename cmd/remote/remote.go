package remotecmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ubclaunchpad/inertia/cmd/inpututil"
	"github.com/ubclaunchpad/inertia/cmd/printutil"
	"github.com/ubclaunchpad/inertia/cmd/util"

	"github.com/ubclaunchpad/inertia/cfg"
	inertiacmd "github.com/ubclaunchpad/inertia/cmd/cmd"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/local"
)

type RemoteCmd struct {
	*cobra.Command
	config  *cfg.Config
	cfgPath string
}

func AttachRemoteCmd(inertia *inertiacmd.Cmd) {
	var remote = RemoteCmd{}
	remote.Command = &cobra.Command{
		Use:   "remote",
		Short: "Configure the local settings for a remote host",
		Long: `Configures local settings for a remote host - add, remove, and list configured
Inertia remotes.

Requires Inertia to be set up via 'inertia init'.

For example:
inertia init
inertia remote add gcloud
inertia gcloud init        # set up Inertia
inertia gcloud status      # check on status of Inertia daemon
`,
		PreRun: func(*cobra.Command, []string) {
			// Ensure project initialized, load config
			var err error
			remote.config, remote.cfgPath, err = local.GetProjectConfigFromDisk(inertia.ConfigPath)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	// add children
	remote.attachAddCmd()
	remote.attachListCmd()
	remote.attachRemoveCmd()
	remote.attachSetCmd()
	remote.attachShowCmd()

	// add to parent
	inertia.AddCommand(remote.Command)
}

func (root *RemoteCmd) attachAddCmd() {
	var addRemote = &cobra.Command{
		Use:   "add [remote]",
		Short: "Add a reference to a remote VPS instance",
		Long: `Adds a reference to a remote VPS instance. Requires information about the VPS
	including IP address, user and a PEM file. The provided name will be used in other
	Inertia commands.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if _, found := root.config.GetRemote(args[0]); found {
				log.Fatal(errors.New("Remote " + args[0] + " already exists."))
			}

			var port, _ = cmd.Flags().GetString("port")
			var sshPort, _ = cmd.Flags().GetString("ssh.port")
			branch, err := local.GetRepoCurrentBranch()
			if err != nil {
				log.Fatal(err)
			}

			// Start prompts and save configuration
			if err = inpututil.AddRemoteWalkthrough(
				os.Stdin, root.config, args[0], port, sshPort, branch,
			); err != nil {
				log.Fatal(err)
			}
			if err = root.config.Write(root.cfgPath); err != nil {
				log.Fatal(err)
			}

			fmt.Println("\nRemote '" + args[0] + "' has been added!")
			fmt.Println("You can now run 'inertia " + args[0] + " init' to set this remote up")
			fmt.Println("for continuous deployment.")
		},
	}
	addRemote.Flags().StringP("port", "p", "4303", "remote daemon port")
	addRemote.Flags().StringP("ssh.port", "s", "22", "remote SSH port")
	root.AddCommand(addRemote)
}

func (root *RemoteCmd) attachListCmd() {
	var list = &cobra.Command{
		Use:   "ls",
		Short: "List currently configured remotes",
		Long:  `Lists all currently configured remotes.`,
		Run: func(cmd *cobra.Command, args []string) {
			var verbose, _ = cmd.Flags().GetBool("verbose")
			for _, remote := range root.config.Remotes {
				if verbose {
					fmt.Println(printutil.FormatRemoteDetails(remote))
				} else {
					fmt.Println(remote.Name)
				}
			}
		},
	}
	list.Flags().BoolP("verbose", "v", false, "enable verbose output")
	root.AddCommand(list)
}

func (root *RemoteCmd) attachRemoveCmd() {
	var remove = &cobra.Command{
		Use:   "rm [remote]",
		Short: "Remove a configured remote",
		Long:  `Remove a remote from Inertia's configuration file.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var _, found = root.config.GetRemote(args[0])
			if found {
				root.config.RemoveRemote(args[0])
				if err := root.config.Write(root.cfgPath); err != nil {
					log.Fatal("Failed to remove remote: " + err.Error())
				}
				fmt.Println("Remote " + args[0] + " removed.")
			} else {
				log.Fatal(errors.New("There does not appear to be a remote with this name. Have you modified the Inertia configuration file?"))
			}
		},
	}
	root.AddCommand(remove)
}

func (root *RemoteCmd) attachShowCmd() {
	var show = &cobra.Command{
		Use:   "show [remote]",
		Short: "Show details about a remote",
		Long:  `Shows details about the given remote.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			remote, found := root.config.GetRemote(args[0])
			if found {
				fmt.Println(printutil.FormatRemoteDetails(remote))
			} else {
				println("No remote '" + args[0] + "' currently set up.")
			}
		},
	}
	root.AddCommand(show)
}

func (root *RemoteCmd) attachSetCmd() {
	var set = &cobra.Command{
		Use:   "set [remote] [property] [value]",
		Short: "Update details about remote",
		Long:  `Updates the given property of the given remote's configuration.`,
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			remote, found := root.config.GetRemote(args[0])
			if found {
				var success = util.SetProperty(args[1], args[2], remote)
				if success {
					root.config.Write(root.cfgPath)
					println("Remote '" + args[0] + "' has been updated.")
					println(printutil.FormatRemoteDetails(remote))
				} else {
					// invalid input
					println("Remote setting '" + args[1] + "' not found.")
				}
			} else {
				println("No remote '" + args[0] + "' currently set up.")
			}
		},
	}
	root.AddCommand(set)
}
