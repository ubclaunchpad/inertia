package remotecmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/input"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
)

// RemoteCmd is the parent class for the 'inertia remote' subcommands
type RemoteCmd struct {
	*cobra.Command
	remotes *cfg.Remotes
}

// AttachRemoteCmd attaches 'remote' subcommands to the given parent command
func AttachRemoteCmd(inertia *core.Cmd) {
	var remote = RemoteCmd{}
	remote.Command = &cobra.Command{
		Use:     "remote",
		Version: inertia.Version,
		Aliases: []string{"r"},
		Short:   "Configure the local settings for a remote host",
		Long: `Configures local settings for a remote host - add, remove, and list configured
Inertia remotes.

Requires Inertia to be set up via 'inertia init'. To see where the remote
configuration is stored, run 'inertia remote config-path'.

For example:

	inertia init
	inertia remote add gcloud
	inertia gcloud init        # set up Inertia
	inertia gcloud status      # check on status of Inertia daemon
`,
		PersistentPreRun: func(*cobra.Command, []string) {
			// Ensure project initialized, load config
			var err error
			remote.remotes, err = local.GetRemotes()
			if err != nil {
				out.Fatal(err)
			}
		},
	}

	// add children
	remote.attachAddCmd()
	remote.attachLoginCmd()
	remote.attachShowCmd()
	remote.attachSetCmd()
	remote.attachListCmd()
	remote.attachRemoveCmd()
	remote.attachUpgradeCmd()
	remote.attachResetCmd()
	remote.attachConfigPathCmd()

	// add to parent
	inertia.AddCommand(remote.Command)
}

func (root *RemoteCmd) validateRemoteName(name string) error {
	if _, found := root.remotes.GetRemote(name); found {
		return fmt.Errorf("remote '%s' already exists", name)
	}
	for _, child := range root.Parent().Commands() {
		if child.Name() == name {
			return fmt.Errorf(
				"'%s' is the name of an Inertia command - please choose something else",
				name)
		}
	}
	return nil
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
			if err := root.validateRemoteName(args[0]); err != nil {
				out.Fatalf(err.Error())
			}

			homeEnvVar, err := os.UserHomeDir()
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

			out.Printf("creating new remote '%s'\n", args[0])
			var highlight = out.NewColorer(out.CY)
			if addr == "" {
				addr, err = input.NewPrompt(nil).
					Prompt(highlight.S(":globe_with_meridians: Enter IP address of remote:")).
					GetString()
				if err != nil {
					out.Fatal(err)
				}
			}

			if keyPath == "" {
				resp, err := input.NewPrompt(&input.PromptConfig{AllowEmpty: true}).
					Prompt(highlight.Sf(":key: Enter path to identity file (leave blank to use '%s'):", defaultKey)).
					GetString()
				if err == nil && resp != "" {
					keyPath = resp
				} else {
					keyPath = defaultKey
				}
			}
			if user == "" {
				user, err = input.NewPrompt(nil).
					Prompt(highlight.Sf(":dancer: Enter user for the identity file:")).
					GetString()
				if err != nil {
					out.Fatal(err)
				}
			}

			var webhookSecret string
			if !genWebhookSecret {
				secret, err := input.NewPrompt(&input.PromptConfig{AllowEmpty: true}).
					Prompt(highlight.Sf(":secret: Enter a webhook secret (leave blank to generate one):")).
					GetString()
				if err == nil && secret != "" {
					webhookSecret = secret
				} else {
					out.Println("generating a webhook secret...")
					webhookSecret, err = common.GenerateRandomString()
					if err != nil {
						out.Fatal(err)
					}
				}
			} else {
				out.Println("generating a webhook secret...")
				webhookSecret, err = common.GenerateRandomString()
				if err != nil {
					out.Fatal(err)
				}
			}

			out.Println("saving new remote...")
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
			out.Printf(highlight.Sf(":boat: Remote '%s' has been added!\n", args[0]))
			out.Printf("You can now run 'inertia %s init' to set this remote up for continuous deployment.\n",
				args[0])
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

func (root *RemoteCmd) attachLoginCmd() {
	const (
		flagIP         = "ip"
		flagTotp       = "totp"
		flagDaemonPort = "daemon.port"
	)
	var loginCmd = &cobra.Command{
		Use:     "login [remote] [user]",
		Short:   "Log in to a remote as an existing user.",
		Long:    `Log in as an existing user to access a remote.`,
		Example: "inertia remote login staging my_user --ip 1.2.3.4",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var (
				remoteName = args[0]
				username   = args[1]
				addr, _    = cmd.Flags().GetString(flagIP)
				port, _    = cmd.Flags().GetString(flagDaemonPort)
			)

			// check that remote name is valid
			if err := root.validateRemoteName(remoteName); err != nil {
				out.Fatal(err.Error())
			}

			// init remote
			remoteCfg := &cfg.Remote{
				Name:     remoteName,
				Version:  root.Version,
				IP:       addr,
				Daemon:   &cfg.Daemon{Port: port},
				Profiles: make(map[string]string),
			}

			// prompt for password
			out.Print("Password: ")
			pwBytes, err := terminal.ReadPassword(int(syscall.Stdin))
			out.Println()
			if err != nil {
				out.Fatal(err.Error())
			}

			// set up client
			var c = client.
				NewClient(remoteCfg, client.Options{Out: os.Stdout}).
				GetUserClient()
			ctx, cancel := context.WithCancel(context.Background())
			input.CatchSigterm(cancel)

			// authenticate
			var totp, _ = cmd.Flags().GetString(flagTotp)
			var req = client.AuthenticateRequest{
				User:     username,
				Password: string(pwBytes),
				TOTP:     totp,
			}
			token, err := c.Authenticate(ctx, req)
			if err != nil && err != client.ErrNeedTotp {
				out.Fatal(err)
			}
			if err == client.ErrNeedTotp {
				// a TOTP is required
				out.Print("Authentication code (or backup code): ")
				totpBytes, err := terminal.ReadPassword(int(syscall.Stdin))
				out.Println()
				if err != nil {
					out.Fatal(err)
				}
				// retry with TOTP
				req.TOTP = string(totpBytes)
				token, err = c.Authenticate(ctx, req)
				if err != nil {
					out.Fatal(err)
				}
			}

			// init token and save remote
			remoteCfg.Daemon.Token = token
			if err := local.SaveRemote(remoteCfg); err != nil {
				out.Fatal(err)
			}
			out.Printf(":rocket: Successfully logged in to remote '%s' (%s) as user '%s'!",
				remoteName, addr, username)
			out.Printf("Try running 'inertia %s status' to check on your remote.", remoteName)
		},
	}
	loginCmd.Flags().String(flagIP, "", "IP address of remote")
	loginCmd.Flags().String(flagTotp, "", "auth code or backup code for 2FA")
	loginCmd.Flags().String(flagDaemonPort, "4303", "remote daemon port")

	root.AddCommand(loginCmd)
}

func (root *RemoteCmd) attachListCmd() {
	const flagVerbose = "verbose"
	var list = &cobra.Command{
		Use:   "ls",
		Short: "List currently configured remotes",
		Long:  `Lists all currently configured remotes.`,
		Run: func(cmd *cobra.Command, args []string) {
			var verbose, _ = cmd.Flags().GetBool(flagVerbose)
			for _, remote := range root.remotes.Remotes {
				if verbose {
					out.Print(out.C("remote '%s'\n", out.BO, out.CY).With(remote.Name))
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
			remote, found := root.remotes.GetRemote(args[0])
			if found {
				out.Print(out.C("remote '%s'\n", out.BO, out.CY).With(remote.Name))
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
				for _, r := range root.remotes.Remotes {
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
					if r, ok := root.remotes.GetRemote(n); ok {
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
			remote, found := root.remotes.GetRemote(args[0])
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

func (root *RemoteCmd) attachResetCmd() {
	var resetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Reset all remotes",
		Long: `Removes all inertia remotes configuration - use 'inertia remote config-path'
to see where the file is directly. Note that the configuration directory can be set using
INERTIA_PATH.`,
		Run: func(cmd *cobra.Command, args []string) {
			should, err := input.NewPrompt(nil).
				Prompt("Would you like to reset ALL remote configuration? (y/N)").
				GetBool()
			if err != nil {
				out.Fatal(err)
			}
			if should {
				if err := os.Remove(local.InertiaRemotesPath()); err != nil {
					out.Fatal(err)
				} else {
					println("remote configuration successfully removed")
				}
			} else {
				out.Fatal("aborting")
			}
		},
	}
	root.AddCommand(resetCmd)
}

func (root *RemoteCmd) attachConfigPathCmd() {
	var cfgPath = &cobra.Command{
		Use:   "config-path",
		Short: "Output path to remote configuration file",
		Long: `Outputs where remotes are stored. Note that the configuration directory
can be set using INERTIA_PATH.`,
		Run: func(cmd *cobra.Command, args []string) {
			out.Println(local.InertiaRemotesPath())
		},
	}
	root.AddCommand(cfgPath)
}
