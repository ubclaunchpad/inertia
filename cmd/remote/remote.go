package remotecmd

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/input"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
)

// RemoteCmd is the parent class for the 'inertia remote' subcommands
type RemoteCmd struct {
	*cobra.Command
	config *cfg.Inertia
}

// AttachRemoteCmd attaches 'remote' subcommands to the given parent command
func AttachRemoteCmd(inertia *core.Cmd) {
	var remote = RemoteCmd{}
	remote.Command = &cobra.Command{
		Use:     "remote",
		Version: inertia.Version,
		Short:   "Configure the local settings for a remote host",
		Long: `Configures local settings for a remote host - add, remove, and list configured
Inertia remotes.

Requires Inertia to be set up via 'inertia init'.

For example:
inertia init
inertia remote add gcloud
inertia gcloud init        # set up Inertia
inertia gcloud status      # check on status of Inertia daemon
`,
		PersistentPreRun: func(*cobra.Command, []string) {
			// Ensure project initialized, load config
			var err error
			remote.config, err = local.GetInertiaConfig()
			if err != nil {
				out.Fatal(err)
			}
		},
	}

	// add children
	remote.attachAddCmd()
	remote.attachShowCmd()
	remote.attachSetCmd()
	remote.attachListCmd()
	remote.attachRemoveCmd()
	remote.attachUpgradeCmd()

	// add to parent
	inertia.AddCommand(remote.Command)
}

func (root *RemoteCmd) attachAddCmd() {
	const (
		flagIP              = "ip"
		flagDaemonPort      = "daemon.port"
		flagSSHPort         = "ssh.port"
		flagSSHKey          = "ssh.key"
		flagSSHUser         = "ssh.user"
		flagWebhookGenerate = "daemon.gen-secret"
	)
	var (
		daemonPort       string
		sshPort          string
		genWebhookSecret bool
	)
	var addRemote = &cobra.Command{
		Use:   "add [remote]",
		Short: "Add a reference to a remote VPS instance",
		Long: `Adds a reference to a remote VPS instance. Requires information about the VPS
including IP address, user and a identity file. The provided name will be used in other
Inertia commands.`,
		Example: "inertia remote add staging --daemon.gen-secret --ip 1.2.3.4",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, child := range root.Parent().Commands() {
				if child.Name() == args[0] {
					out.Fatalf("'%s' is the name of an Inertia command - please choose something else", args[0])
				}
			}
			if _, found := root.config.GetRemote(args[0]); found {
				out.Fatalf("remote '%s' already exists", args[0])
			}
			homeEnvVar, err := local.GetHomePath()
			if err != nil {
				out.Fatal(err)
			}

			var (
				sshDir     = filepath.Join(homeEnvVar, ".ssh")
				defaultKey = filepath.Join(sshDir, "id_rsa")
				addr, _    = cmd.Flags().GetString(flagIP)
				port, _    = cmd.Flags().GetString(flagDaemonPort)
				sshPort, _ = cmd.Flags().GetString(flagSSHPort)
				keyPath, _ = cmd.Flags().GetString(flagSSHKey)
				user, _    = cmd.Flags().GetString(flagSSHUser)
			)

			if addr == "" {
				addr, err = input.Prompt("Enter IP address of remote:")
				if err != nil {
					out.Fatal(err)
				}
			}

			if keyPath == "" {
				if resp, err := input.Promptf(
					"Enter location of identity file (leave blank to use '%s'):", defaultKey,
				); err == nil && resp != "" {
					keyPath = resp
				} else {
					keyPath = defaultKey
				}
			}
			if user == "" {
				user, err = input.Prompt("Enter user:")
				if err != nil {
					out.Fatal(err)
				}
			}

			var webhookSecret string
			if !genWebhookSecret {
				secret, err := input.Prompt("Enter webhook secret (leave blank to generate one):")
				if err == nil && secret != "" {
					webhookSecret = secret
				} else {
					webhookSecret, err = common.GenerateRandomString()
					if err != nil {
						out.Fatal(err)
					}
				}
			} else {
				webhookSecret, err = common.GenerateRandomString()
				if err != nil {
					out.Fatal(err)
				}
			}

			if err := local.SaveRemote(&cfg.Remote{
				Name:    args[0],
				Version: root.Version,
				IP:      addr,
				SSH: &cfg.SSH{
					User:         user,
					IdentityFile: keyPath,
					SSHPort:      sshPort,
				},
				Daemon: &cfg.Daemon{
					Port:          port,
					WebHookSecret: webhookSecret,
				},
				Profiles: make(map[string]string),
			}); err != nil {
				out.Fatal(err)
			}

			out.Printf(`
Remote '%s' has been added!
You can now run 'inertia %s init' to set this remote up for continuous deployment.
`, args[0], args[0])
		},
	}
	addRemote.Flags().String(flagIP, "", "IP address of remote")
	addRemote.Flags().StringVar(&daemonPort, flagDaemonPort, "4303", "remote daemon port")
	addRemote.Flags().StringVar(&sshPort, flagSSHPort, "22", "remote SSH port")
	addRemote.Flags().String(flagSSHKey, "", "path to SSH key for remote")
	addRemote.Flags().String(flagSSHUser, "", "user to use when accessing remote over SSH")
	addRemote.Flags().BoolVar(&genWebhookSecret, flagWebhookGenerate, true, "toggle webhook secret generation")
	root.AddCommand(addRemote)
}

func (root *RemoteCmd) attachListCmd() {
	const flagVerbose = "verbose"
	var list = &cobra.Command{
		Use:   "ls",
		Short: "List currently configured remotes",
		Long:  `Lists all currently configured remotes.`,
		Run: func(cmd *cobra.Command, args []string) {
			var verbose, _ = cmd.Flags().GetBool(flagVerbose)
			for _, remote := range root.config.Remotes {
				if verbose {
					out.Printf("remote '%s'\n", remote.Name)
					out.Println(out.FormatRemoteDetails(*remote))
				} else {
					out.Println(remote.Name)
				}
			}
		},
	}
	list.Flags().BoolP(flagVerbose, "v", false, "enable verbose out")
	root.AddCommand(list)
}

func (root *RemoteCmd) attachRemoveCmd() {
	var remove = &cobra.Command{
		Use:     "rm [remote]",
		Short:   "Remove a configured remote",
		Long:    `Remove a remote from Inertia's configuration file.`,
		Example: "inertia remote rm staging",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			out.Printf("removing remotes %s\n", strings.Join(args, ", "))
			for _, r := range args {
				if err := local.RemoveRemote(r); err != nil {
					out.Fatal(err.Error())
				} else {
					out.Printf("remote '%s' removed\n", r)
				}
			}
		},
	}
	root.AddCommand(remove)
}

func (root *RemoteCmd) attachShowCmd() {
	var show = &cobra.Command{
		Use:     "show [remote]",
		Short:   "Show details about a remote",
		Long:    `Shows details about the given remote.`,
		Example: "inertia remote show staging",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			remote, found := root.config.GetRemote(args[0])
			if found {
				out.Println(out.FormatRemoteDetails(*remote))
			} else {
				out.Println("no remote '" + args[0] + "' currently configured")
			}
		},
	}
	root.AddCommand(show)
}

func (root *RemoteCmd) attachUpgradeCmd() {
	const (
		flagVersion = "version"
		flagAll     = "all"
	)
	var upgrade = &cobra.Command{
		Use:     "upgrade",
		Short:   "Upgrade your remote configuration version to match the CLI",
		Long:    `Upgrade your remote configuration version to match the CLI and save it to global settings.`,
		Example: "inertia remote upgrade dev staging",
		Run: func(cmd *cobra.Command, args []string) {
			var version = root.Version
			if v, _ := cmd.Flags().GetString(flagVersion); v != "" {
				version = v
			}

			var all, _ = cmd.Flags().GetBool(flagAll)
			if (len(args) == 0) && !all {
				cmd.Help()
				out.Println()
				out.Fatal("you must provide remotes or use the '--all' flag")
			}

			var remotes = args
			if all {
				out.Printf("updating configuration to version '%s' for all remotes\n", version)
				for _, r := range root.config.Remotes {
					r.Version = version
					if err := local.SaveRemote(r); err != nil {
						out.Fatalf("could not update remote '%s': %s", r.Name, err.Error())
					} else {
						out.Printf("remote '%s' updated\n", r.Name)
					}
				}
			} else {
				out.Printf("setting configuration to version '%s' for remotes %s\n",
					version, strings.Join(remotes, ", "))
				for _, n := range remotes {
					if r, ok := root.config.GetRemote(n); ok {
						r.Version = version
						if err := local.SaveRemote(r); err != nil {
							out.Fatalf("could not update remote '%s': %s", n, err.Error())
						} else {
							out.Printf("remote '%s' updated\n", n)
						}
					} else {
						out.Fatalf("could not find remote '%s'", n)
					}
				}
			}
		},
	}
	upgrade.Flags().Bool(flagAll, false, "upgrade all remotes")
	upgrade.Flags().String(flagVersion, root.Version, "specify Inertia daemon version to set")
	root.AddCommand(upgrade)
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
				if err := cfg.SetProperty(args[1], args[2], remote); err == nil {
					if err := local.SaveRemote(remote); err != nil {
						out.Fatal(err.Error())
					}
					out.Println("remote '" + args[0] + "' has been updated")
				} else {
					out.Fatalf("could not update remote '%s': %s", args[0], err.Error())
				}
			} else {
				out.Println("No remote '" + args[0] + "' currently set up.")
			}
		},
	}
	root.AddCommand(set)
}
