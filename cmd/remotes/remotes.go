package remotescmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/client/bootstrap"
	"github.com/ubclaunchpad/inertia/client/runner"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/input"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
)

// AttachRemotesCmds reads configuration to attach a child command for each
// configured remote in the configuration
func AttachRemotesCmds(root *core.Cmd, validateConfig bool) {
	project, _ := local.GetProject(root.ProjectConfigPath)

	// execute some project validation, since AttachRemotesCmds always runs
	if project != nil && validateConfig {
		if project.InertiaMinVersion == "" {
			// init version if none is set
			project.InertiaMinVersion = root.Version
			local.Write(root.ProjectConfigPath, project)
		} else {
			// else validate the provided version
			var msg string
			warn, err := project.ValidateVersion(root.Version)
			if err != nil || warn != "" {
				if err != nil {
					msg += out.C(":warning: error when validating project configuration against CLI version: %s\n",
						out.RD, out.BO).With(err).String()
				}
				if warn != "" {
					msg += out.C(":warning: warning when validating project configuration against CLI version: %s\n",
						out.YE, out.BO).With(warn).String()
				}
				msg += out.C("for details on the latest Inertia releases, please see https://github.com/ubclaunchpad/inertia/releases/latest\n",
					out.BO).String()
			}
			out.Println(msg)
		}
	}

	// parse and attach remotes
	cfg, err := local.GetRemotes()
	if err != nil {
		return
	}
	var remotes = make(map[string]bool)
	for _, r := range cfg.Remotes {
		if _, ok := remotes[r.Name]; ok {
			out.Fatalf("you have configured multiple remotes named '%s' - please rename one in %s",
				r.Name, local.InertiaRemotesPath())
		}
		for _, child := range root.Commands() {
			if child.Name() == r.Name {
				out.Fatalf("you have configured a remote named '%s', which is an Inertia command - please rename it in %s",
					r.Name, local.InertiaRemotesPath())
			}
		}
		remotes[r.Name] = true
		AttachRemoteHostCmd(root, CmdOptions{
			RemoteCfg:  r,
			ProjectCfg: project,
		})
	}
}

// HostCmd is the parent class for a subcommand for a configured remote host
type HostCmd struct {
	*cobra.Command
	remote  string
	project *cfg.Project

	client *client.Client
	ctx    context.Context
}

// CmdOptions denotes options for individual host subcommands
type CmdOptions struct {
	RemoteCfg  *cfg.Remote
	ProjectCfg *cfg.Project
}

const (
	flagShort = "short"
	flagDebug = "debug"
)

// AttachRemoteHostCmd attaches a subcommand for a configured remote host to the
// given parent
func AttachRemoteHostCmd(
	inertia *core.Cmd,
	opts CmdOptions,
	hidden ...bool,
) {
	ctx, cancel := context.WithCancel(context.Background())
	input.CatchSigterm(cancel)
	var host = &HostCmd{
		project: opts.ProjectCfg,
		client: client.NewClient(opts.RemoteCfg, client.Options{
			SSH: runner.SSHOptions{
				KeyPassphrase: os.Getenv(local.EnvSSHPassphrase),
			},
			Out: os.Stdout,
		}),
		ctx: ctx,
	}
	host.Command = &cobra.Command{
		Use: opts.RemoteCfg.Name + " [command]",
		Hidden: func() bool {
			// hide command by default
			if len(hidden) > 0 {
				return hidden[0]
			}
			return true
		}(),
		Short: "Configure deployment to " + opts.RemoteCfg.Name,
		Long: `Manages deployment on specified remote.

Requires:
1. an Inertia daemon running on your remote - use 'inertia [remote] init' to get it running.
2. a deploy key to be registered within your remote repository for the daemon to use.

Continuous deployment requires the daemon's webhook address to be registered in your remote repository.

If the SSH key for your remote requires a passphrase, it can be provided via 'IDENTITY_PASSPHRASE'.

Run 'inertia [remote] init' to gather this information.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if host.client == nil {
				out.Fatal("failed to read configuration")
			}
			if host.project == nil {
				out.Fatal("no project found in current directory - try 'inertia init'")
			}
			if host.getRemote().Version != inertia.Version {
				out.Printf("[WARNING] Remote configuration version '%s' does not match your Inertia CLI version '%s'\n",
					host.getRemote().Version, inertia.Version)
			}
			var debug, _ = cmd.Flags().GetBool(flagDebug)
			host.client.WithDebug(debug)
		},
	}
	host.PersistentFlags().BoolP(flagShort, "s", false,
		"don't stream output from command")
	host.PersistentFlags().Bool(flagDebug, false,
		"enable debug output from Inertia client")

	// attach children
	host.attachInitCmd()
	host.attachUpCmd()
	host.attachDownCmd()
	host.attachStatusCmd()
	host.attachLogsCmd()
	AttachUserCmd(host)
	AttachEnvCmd(host)
	host.attachSendFileCmd()
	host.attachSSHCmd()
	host.attachPruneCmd()
	host.attachTokenCmd()
	host.attachUpgradeCmd()
	host.attachUninstallCmd()

	// attach to parent
	inertia.AddCommand(host.Command)
}

func (root *HostCmd) getRemote() *cfg.Remote { return root.client.Remote }

func (root *HostCmd) attachUpCmd() {
	const (
		flagProfile = "profile"
	)
	var up = &cobra.Command{
		Use:   "up",
		Short: "Bring project online on remote",
		Long: `Builds and deploy your project on your remote using your project's
default profile, or a profile you have applied using 'inertia project profile apply'.

This requires an Inertia daemon to be active on your remote - do this by running
'inertia [remote] init'.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Get flags and profile
			var short, _ = cmd.Flags().GetBool(flagShort)
			var profileName = root.getRemote().GetProfile(root.project.Name)
			profile, found := root.project.GetProfile(profileName)
			if !found {
				out.Fatalf("could not find profile '%s'", profileName)
			}
			out.Printf("deploying project '%s' using profile '%s'\n", root.project.Name, profileName)

			// Make up request
			var req = client.UpRequest{
				Project: root.project.Name,
				URL:     root.project.URL,
				Profile: *profile}

			var err error
			if short {
				err = root.client.Up(root.ctx, req)
			} else {
				err = root.client.UpWithOutput(root.ctx, req)
			}
			if err != nil {
				out.Fatal(err)
			}
			if !short {
				out.Println("project deployment successfully started!")
			}
		},
	}
	up.Flags().StringP(flagProfile, "p", "", "specify a profile to deploy")
	root.AddCommand(up)
}

func (root *HostCmd) attachDownCmd() {
	var down = &cobra.Command{
		Use:   "down",
		Short: "Bring project offline on remote",
		Long: `Stops your project on your remote. This will kill all active project containers on your remote.
	
Requires project to be online - do this by running 'inertia [remote] up`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := root.client.Down(root.ctx); err != nil {
				out.Fatal(err)
			}
			out.Println("project successfully shut down")
		},
	}
	root.AddCommand(down)
}

func (root *HostCmd) attachStatusCmd() {
	var stat = &cobra.Command{
		Use:   "status",
		Short: "Print the status of the deployment on this remote",
		Long: `Prints the status of the deployment on this remote.

Requires the Inertia daemon to be active on your remote - do this by running 'inertia [remote] up'`,
		Run: func(cmd *cobra.Command, args []string) {
			status, err := root.client.Status(root.ctx)
			if err != nil {
				out.Fatal(err)
			}

			host, err := root.getRemote().DaemonAddr()
			if err != nil {
				out.Fatal(err)
			}
			out.Printf("daemon on remote '%s' is online at %s\n",
				root.remote, host)
			out.Println(out.FormatStatus("robert", status))
		},
	}
	root.AddCommand(stat)
}

func (root *HostCmd) attachLogsCmd() {
	const flagEntries = "entries"
	var log = &cobra.Command{
		Use:   "logs [container]",
		Short: "Access logs of containers on your remote host",
		Long: `Accesses logs of containers on your remote host.
	
By default, this command retrieves Inertia daemon logs, but you can provide an
argument that specifies the name of the container you wish to retrieve logs for.
Use 'inertia [remote] status' to see which containers are active.`,
		Run: func(cmd *cobra.Command, args []string) {
			var short, _ = cmd.Flags().GetBool(flagShort)
			var entries, _ = cmd.Flags().GetInt(flagEntries)

			// get daemon logs by default
			var container = "/inertia-daemon"
			if len(args) > 0 {
				container = args[0]
			}

			var req = client.LogsRequest{
				Container: container,
				Entries:   entries}

			if short {
				// if short, just grab the last x log entries
				logs, err := root.client.Logs(root.ctx, req)
				if err != nil {
					out.Fatal(err)
				}
				out.Println(strings.Join(logs, "\n"))
			} else {
				// if not short, open a websocket to stream logs
				if err := root.client.LogsWithOutput(root.ctx, req); err != nil {
					out.Fatal(err)
				}
			}
		},
	}
	log.Flags().Int(flagEntries, 0, "Number of log entries to fetch")
	root.AddCommand(log)
}

func (root *HostCmd) attachPruneCmd() {
	var prune = &cobra.Command{
		Use:   "prune",
		Short: "Prune Docker assets and images on your remote",
		Long:  `Prunes Docker assets and images from your remote to free up storage space.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := root.client.Prune(root.ctx); err != nil {
				out.Fatal(err)
			}
			out.Printf("docker assets have been pruned")
		},
	}
	root.AddCommand(prune)
}

func (root *HostCmd) attachSSHCmd() {
	var ssh = &cobra.Command{
		Use:   "ssh",
		Short: "Start an interactive SSH session",
		Long:  `Starts an interact SSH session with your remote.`,
		Run: func(cmd *cobra.Command, args []string) {
			sshc, err := root.client.GetSSHClient()
			if err != nil {
				out.Fatal(err.Error())
			}
			if err := sshc.GetRunner().RunSession(args...); err != nil {
				out.Fatal(err.Error())
			}
		},
	}
	root.AddCommand(ssh)
}

func (root *HostCmd) attachSendFileCmd() {
	const (
		flagDest = "dest"
		flagPerm = "perm"
	)
	var sendFile = &cobra.Command{
		Use:   "send [filepath]",
		Short: "Send a file to your Inertia deployment",
		Long:  `Sends a file, such as a configuration or .env file, to your Inertia deployment.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			// Get permissions to copy file with
			var permissions, _ = cmd.Flags().GetString(flagPerm)

			// Open file with given name
			cwd, err := os.Getwd()
			if err != nil {
				out.Fatal(err.Error())
			}
			f, err := os.Open(path.Join(cwd, args[0]))
			if err != nil {
				out.Fatal(err.Error())
			}

			// Get flag for destination
			var dest, _ = cmd.Flags().GetString(flagDest)
			if dest == "" {
				dest = args[0]
			}

			// Destination path - todo: allow config
			var projectPath = "$HOME/inertia/project"
			var remotePath = path.Join(projectPath, dest)

			// Initiate copy
			sshc, err := root.client.GetSSHClient()
			if err != nil {
				out.Fatal(err.Error())
			}
			if err = sshc.GetRunner().CopyFile(f, remotePath, permissions); err != nil {
				out.Fatal(err.Error())
			}

			out.Println("File", args[0], "has been copied to", remotePath, "on remote", root.remote)
		},
	}
	sendFile.Flags().StringP(flagDest, "d", "", "path relative from project root to send file to")
	sendFile.Flags().StringP(flagPerm, "p", "0655", "permissions settings to create file with")
	root.AddCommand(sendFile)
}

func (root *HostCmd) attachInitCmd() {
	var init = &cobra.Command{
		Use:   "init",
		Short: "Initialize remote host for deployment",
		Long: `Initializes this remote host for deployment.

This command sets up your remote host and brings an Inertia daemon online on your remote.

Upon successful setup, you will be provided with:
	- a deploy key
	- a webhook URL

The deploy key is required for the daemon to access your repository, and the
webhook URL enables continuous deployment as your repository is updated.`,
		Run: func(cmd *cobra.Command, args []string) {
			var repo = common.ExtractRepository(common.GetSSHRemoteURL(root.project.URL))
			if err := bootstrap.Bootstrap(root.client, bootstrap.Options{
				RepoName: repo,
				Out:      os.Stdout,
			}); err != nil {
				out.Fatal(err.Error())
			}

			// write back to configuration
			if err := local.SaveRemote(root.getRemote()); err != nil {
				out.Fatal(err)
			}
		},
	}
	root.AddCommand(init)
}

func (root *HostCmd) newResetCmd() {
	var reset = &cobra.Command{
		Use:   "reset",
		Short: "Reset the project on your remote",
		Long: `Resets the project on your remote.

On this remote, this kills all active containers and clears the project directory,
allowing you to assign a different Inertia project to this remote.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := root.client.Reset(root.ctx); err != nil {
				out.Fatal(err)
			}
			out.Printf("project on remote '%s' successfully reset\n", root.remote)
		},
	}
	root.AddCommand(reset)
}

func (root *HostCmd) attachUninstallCmd() {
	var uninstall = &cobra.Command{
		Use:   "uninstall",
		Short: "Shut down Inertia and remove Inertia assets from remote host",
		Long: `Shuts down and removes the Inertia daemon, and removes the Inertia
directory (~/inertia) from your remote host.`,
		Run: func(cmd *cobra.Command, args []string) {
			out.Println("WARNING: This will stop down your project and remove the Inertia daemon.")
			out.Println("This is irreversible. Continue? (y/n)")
			var response string
			if _, err := fmt.Scanln(&response); err != nil || response != "y" {
				out.Fatal("aborting")
			}

			sshc, err := root.client.GetSSHClient()
			if err != nil {
				out.Fatal(err)
			}

			// Daemon down
			out.Println("Stopping project...")
			if err = root.client.Down(root.ctx); err != nil {
				out.Fatal(err)
			}
			out.Println("Stopping daemon...")
			if err = sshc.DaemonDown(); err != nil {
				out.Fatal(err)
			}
			out.Println("Removing Inertia directories...")
			if err = sshc.UninstallInertia(); err != nil {
				out.Fatal(err)
			}
			out.Println("Uninstallation completed.")
		},
	}
	root.AddCommand(uninstall)
}

func (root *HostCmd) attachTokenCmd() {
	var tokenCmd = &cobra.Command{
		Use:   "token",
		Short: "Generate tokens associated with permission levels for admin to share.",
		Long:  `Generate tokens associated with permission levels for team leads to share`,
		Run: func(cmd *cobra.Command, args []string) {
			useSSH, err := cmd.Flags().GetBool("ssh")
			if err != nil {
				out.Fatal(err)
			}
			var token string
			if useSSH {
				sshc, err := root.client.GetSSHClient()
				if err != nil {
					out.Fatal(err.Error())
				}
				if err = sshc.AssignAPIToken(); err != nil {
					out.Fatal(err.Error())
				}
				token = root.client.Remote.Daemon.Token
			} else {
				token, err = root.client.Token(root.ctx)
				if err != nil {
					out.Fatal(err)
				}
			}
			out.Println(token)
		},
	}
	tokenCmd.Flags().Bool("ssh", false, "generate token over SSH")
	root.AddCommand(tokenCmd)
}

func (root *HostCmd) attachUpgradeCmd() {
	const (
		flagVersion = "version"
	)
	var upgrade = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Inertia daemon to match the CLI.",
		Long: `Restarts the Inertia daemon to upgrade it to the same version as your CLI.

To upgrade your remote, you must upgrade your CLI first to the correct version - drop by
https://github.com/ubclaunchpad/inertia/releases/latest for more details.`,
		Run: func(cmd *cobra.Command, args []string) {
			sshc, err := root.client.GetSSHClient()
			if err != nil {
				out.Fatal(err)
			}

			out.Println("Shutting down daemon...")
			if err = sshc.DaemonDown(); err != nil {
				out.Fatal(err)
			}

			if v, _ := cmd.Flags().GetString(flagVersion); v != "" {
				root.getRemote().Version = v
			}

			out.Printf("Starting up the Inertia daemon (version %s)\n", root.getRemote().Version)
			if err := sshc.DaemonUp(); err != nil {
				out.Fatal(err)
			}

			if err := local.SaveRemote(root.getRemote()); err != nil {
				out.Fatal(err)
			}
		},
	}
	upgrade.Flags().String(flagVersion, "", "version of Inertia daemon to spin up")
	root.AddCommand(upgrade)
}
