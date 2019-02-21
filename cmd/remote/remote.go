package remotecmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/input"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/output"
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
				output.Fatal(err)
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
including IP address, user and a PEM file. The provided name will be used in other
Inertia commands.`,
		Example: "inertia remote add staging --daemon.gen-secret --ip 1.2.3.4",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if _, found := root.config.GetRemote(args[0]); found {
				output.Fatal(errors.New("remote " + args[0] + " already exists"))
			}
			homeEnvVar, err := local.GetHomePath()
			if err != nil {
				output.Fatal(err)
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
					output.Fatal(err)
				}
			}

			if keyPath == "" {
				if resp, err := input.Promptf(
					"Enter location of PEM file (leave blank to use '%s'):", defaultKey,
				); err == nil && resp != "" {
					keyPath = resp
				} else {
					keyPath = defaultKey
				}
			}
			if user == "" {
				user, err = input.Prompt("Enter user:")
				if err != nil {
					output.Fatal(err)
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
						output.Fatal(err)
					}
				}
			} else {
				webhookSecret, err = common.GenerateRandomString()
				if err != nil {
					output.Fatal(err)
				}
			}

			if err := local.SaveRemote(args[0], &cfg.Remote{
				Version: root.Version,
				IP:      addr,
				SSH: &cfg.SSH{
					User:    user,
					PEM:     keyPath,
					SSHPort: sshPort,
				},
				Daemon: &cfg.Daemon{
					Port:          port,
					WebHookSecret: webhookSecret,
				},
				Profiles: make(map[string]string),
			}); err != nil {
				output.Fatal(err)
			}

			fmt.Printf(`
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
			for name, remote := range root.config.Remotes {
				if verbose {
					fmt.Println(output.FormatRemoteDetails(name, remote))
				} else {
					fmt.Println(name)
				}
			}
		},
	}
	list.Flags().BoolP(flagVerbose, "v", false, "enable verbose output")
	root.AddCommand(list)
}

func (root *RemoteCmd) attachRemoveCmd() {
	var remove = &cobra.Command{
		Use:     "rm [remote]",
		Short:   "Remove a configured remote",
		Long:    `Remove a remote from Inertia's configuration file.`,
		Example: "inertia remote rm staging",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := local.SaveRemote(args[0], nil); err != nil {
				output.Fatal(err.Error())
			} else {
				fmt.Printf("remote '%s' removed\n", args[0])
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
				fmt.Println(output.FormatRemoteDetails(args[0], *remote))
			} else {
				println("no remote '" + args[0] + "' currently configured")
			}
		},
	}
	root.AddCommand(show)
}

func (root *RemoteCmd) attachUpgradeCmd() {
	const flagVersion = "version"
	const flagRemotes = "remotes"
	var upgrade = &cobra.Command{
		Use:     "upgrade",
		Short:   "Upgrade your remote configuration version to match the CLI",
		Long:    `Upgrade your remote configuration version to match the CLI and save it to global settings.`,
		Example: "inertia remote upgrade -r dev -r staging",
		Run: func(cmd *cobra.Command, args []string) {
			// Ensure project initialized.
			config, err := local.GetInertiaConfig()
			if err != nil {
				output.Fatal(err)
			}

			var version = root.Version
			if v, _ := cmd.Flags().GetString(flagVersion); v != "" {
				version = v
			}

			var remotes, _ = cmd.Flags().GetStringArray(flagRemotes)
			if len(remotes) == 0 {
				fmt.Printf("updating configuration to version '%s' for all remotes", version)
				for n, r := range config.Remotes {
					r.Version = version
					if err := local.SaveRemote(n, &r); err != nil {
						output.Fatalf("could not update remote '%s': %s", n, err.Error())
					} else {
						fmt.Printf("remote '%s' updated\n", n)
					}
				}
			} else {
				fmt.Printf("setting configuration to version '%s' for remotes %s\n",
					version, strings.Join(remotes, ", "))
				for _, n := range remotes {
					if r, ok := config.Remotes[n]; ok {
						r.Version = version
						if err := local.SaveRemote(n, &r); err != nil {
							output.Fatalf("could not update remote '%s': %s", n, err.Error())
						} else {
							fmt.Printf("remote '%s' updated\n", n)
						}
					} else {
						output.Fatalf("could not find remote '%s'", n)
					}
				}
			}
		},
	}
	upgrade.Flags().String(flagVersion, root.Version, "specify Inertia daemon version to set")
	upgrade.Flags().StringArrayP(flagRemotes, "r", nil, "specify which remotes to modify (default: all)")
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
					local.SaveRemote(args[0], remote)
					println("remote '" + args[0] + "' has been updated")
				} else {
					output.Fatalf("could not update remote '%s': %s", args[0], err.Error())
				}
			} else {
				println("No remote '" + args[0] + "' currently set up.")
			}
		},
	}
	root.AddCommand(set)
}
