package remotescmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/client/bootstrap"
	"github.com/ubclaunchpad/inertia/client/runner"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/output"
	"github.com/ubclaunchpad/inertia/local"
)

const (
	// EnvSSHPassphrase is the key used to fetch PEM key passphrases
	EnvSSHPassphrase = "PEM_PASSPHRASE"
)

// AttachRemotesCmds reads configuration to attach a child command for each
// configured remote in the configuration
func AttachRemotesCmds(root *core.Cmd) {
	project, err := local.GetProject(root.ProjectConfigPath)
	if err != nil {
		return
	}
	cfg, err := local.GetInertiaConfig()
	if err != nil {
		return
	}
	for _, r := range cfg.Remotes {
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
}

// CmdOptions denotes options for individual host subcommands
type CmdOptions struct {
	RemoteCfg  *cfg.Remote
	ProjectCfg *cfg.Project
}

const (
	flagShort = "short"
)

// AttachRemoteHostCmd attaches a subcommand for a configured remote host to the
// given parent
func AttachRemoteHostCmd(
	inertia *core.Cmd,
	opts CmdOptions,
	hidden ...bool,
) {
	var host = &HostCmd{
		project: opts.ProjectCfg,
		client: client.NewClient(opts.RemoteCfg, client.Options{
			SSH: runner.SSHOptions{
				KeyPassphrase: os.Getenv(EnvSSHPassphrase),
			},
			Out: os.Stdout,
		}),
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

If the SSH key for your remote requires a passphrase, it can be provided via 'PEM_PASSPHRASE'.

Run 'inertia [remote] init' to gather this information.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if host.client == nil || host.project == nil {
				output.Fatal("failed to read configuration")
			}
			if host.getRemote().Version != inertia.Version {
				fmt.Printf("[WARNING] Remote configuration version '%s' does not match your Inertia CLI version '%s'\n",
					host.getRemote().Version, inertia.Version)
			}
		},
	}
	host.PersistentFlags().BoolP(flagShort, "s", false,
		"don't stream output from command")

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

func (root *HostCmd) getRemote() *cfg.Remote { return root.getRemote() }

func (root *HostCmd) attachUpCmd() {
	var up = &cobra.Command{
		Use:   "up",
		Short: "Bring project online on remote",
		Long: `Builds and deploy your project on your remote.

This requires an Inertia daemon to be active on your remote - do this by running 'inertia [remote] init'`,
		Run: func(cmd *cobra.Command, args []string) {
			// Get flags
			var short, _ = cmd.Flags().GetBool(flagShort)
			var project = "" // TODO
			profile, found := root.project.GetProfile(root.getRemote().GetProfile(project))
			if !found {
				output.Fatalf("could not find profile '%s'", root.getRemote().GetProfile(project))
			}

			resp, err := root.client.Up(
				project,
				root.project.URL,
				*profile,
				!short)
			if err != nil {
				output.Fatal(err)
			}
			defer resp.Body.Close()

			if short {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					output.Fatal(err)
				}
				switch resp.StatusCode {
				case http.StatusCreated:
					fmt.Printf("(Status code %d) Project build started!\n", resp.StatusCode)
				case http.StatusUnauthorized:
					fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
				case http.StatusPreconditionFailed:
					fmt.Printf("(Status code %d) Problem with deployment setup:\n%s\n", resp.StatusCode, body)
				default:
					fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
						resp.StatusCode, body)
				}
			} else {
				reader := bufio.NewReader(resp.Body)
				for {
					line, err := reader.ReadBytes('\n')
					if err != nil {
						break
					}
					fmt.Print(string(line))
				}
			}
		},
	}
	root.AddCommand(up)
}

func (root *HostCmd) attachDownCmd() {
	var down = &cobra.Command{
		Use:   "down",
		Short: "Bring project offline on remote",
		Long: `Stops your project on your remote. This will kill all active project containers on your remote.
	
Requires project to be online - do this by running 'inertia [remote] up`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := root.client.Down()
			if err != nil {
				output.Fatal(err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				output.Fatal(err)
			}

			switch resp.StatusCode {
			case http.StatusOK:
				fmt.Printf("(Status code %d) Project down\n", resp.StatusCode)
			case http.StatusPreconditionFailed:
				fmt.Printf("(Status code %d) No containers are currently active\n", resp.StatusCode)
			case http.StatusUnauthorized:
				fmt.Printf("(Status code %d) Bad auth: %s\n", resp.StatusCode, body)
			default:
				fmt.Printf("(Status code %d) Unknown response from daemon: %s\n",
					resp.StatusCode, body)
			}
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
			resp, err := root.client.Status()
			if err != nil {
				output.Fatal(err)
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK:
				host, err := root.getRemote().DaemonAddr()
				if err != nil {
					output.Fatal(err)
				}
				fmt.Printf("(Status code %d) Daemon at remote '%s' online at %s\n",
					resp.StatusCode, root.remote, host)
				var status = &api.DeploymentStatus{}
				if _, err := api.Unmarshal(resp.Body, api.KV{
					Key: "status", Value: status,
				}); err != nil {
					output.Fatal(err)
				}
				println(output.FormatStatus(status))
			case http.StatusUnauthorized:
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					output.Fatal(err)
				}
				fmt.Printf("(Status code %d) Bad auth: %s\n", resp.StatusCode, body)
			default:
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					output.Fatal(err)
				}
				fmt.Printf("(Status code %d) %s\n",
					resp.StatusCode, body)
			}
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

			if short {
				// if short, just grab the last x log entries
				resp, err := root.client.Logs(container, entries)
				if err != nil {
					output.Fatal(err)
				}
				defer resp.Body.Close()

				var logs []string
				b, err := api.Unmarshal(resp.Body, api.KV{Key: "logs", Value: &logs})
				if err != nil {
					output.Fatal(err)
				}

				switch resp.StatusCode {
				case http.StatusOK:
					fmt.Printf("(Status code %d) Logs: \n%s\n", resp.StatusCode, strings.Join(logs, "\n"))
				case http.StatusUnauthorized:
					fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, b.Message)
				case http.StatusPreconditionFailed:
					fmt.Printf("(Status code %d) Problem with deployment setup:\n%s\n", resp.StatusCode, b.Message)
				default:
					fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
						resp.StatusCode, b.Message)
				}
			} else {
				// if not short, open a websocket to stream logs
				socket, err := root.client.LogsWebSocket(container, entries)
				if err != nil {
					output.Fatal(err)
				}
				defer socket.Close()

				for {
					_, line, err := socket.ReadMessage()
					if err != nil {
						output.Fatal(err)
					}
					fmt.Print(string(line))
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
			resp, err := root.client.Prune()
			if err != nil {
				output.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				output.Fatal(err)
			}
			fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
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
				output.Fatal(err.Error())
			}
			if err := sshc.GetRunner().RunSession(args...); err != nil {
				output.Fatal(err.Error())
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
				output.Fatal(err.Error())
			}
			f, err := os.Open(path.Join(cwd, args[0]))
			if err != nil {
				output.Fatal(err.Error())
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
				output.Fatal(err.Error())
			}
			if err = sshc.GetRunner().CopyFile(f, remotePath, permissions); err != nil {
				output.Fatal(err.Error())
			}

			fmt.Println("File", args[0], "has been copied to", remotePath, "on remote", root.remote)
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
			if err := bootstrap.SetUpRemote(os.Stdout, root.remote, root.project.URL, root.client); err != nil {
				output.Fatal(err.Error())
			}

			// write back to configuration
			if err := local.SaveRemote(root.getRemote()); err != nil {
				output.Fatal(err)
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
			resp, err := root.client.Reset()
			if err != nil {
				output.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				output.Fatal(err)
			}

			switch resp.StatusCode {
			case http.StatusOK:
				fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
			case http.StatusUnauthorized:
				fmt.Printf("(Status code %d) Bad auth: %s\n", resp.StatusCode, body)
			default:
				fmt.Printf("(Status code %d) Unknown response from daemon: %s\n",
					resp.StatusCode, body)
			}
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
			println("WARNING: This will stop down your project and remove the Inertia daemon.")
			println("This is irreversible. Continue? (y/n)")
			var response string
			_, err := fmt.Scanln(&response)
			if err != nil || response != "y" {
				output.Fatal("aborting")
			}

			sshc, err := root.client.GetSSHClient()
			if err != nil {
				output.Fatal(err)
			}

			// Daemon down
			println("Stopping project...")
			if _, err = root.client.Down(); err != nil {
				output.Fatal(err)
			}
			println("Stopping daemon...")
			if err = sshc.DaemonDown(); err != nil {
				output.Fatal(err)
			}
			println("Removing Inertia directories...")
			if err = sshc.UninstallInertia(); err != nil {
				output.Fatal(err)
			}
			println("Uninstallation completed.")
		},
	}
	root.AddCommand(uninstall)
}

func (root *HostCmd) attachTokenCmd() {
	var token = &cobra.Command{
		Use:   "token",
		Short: "Generate tokens associated with permission levels for admin to share.",
		Long:  `Generate tokens associated with permission levels for team leads to share`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := root.client.Token()
			if err != nil {
				output.Fatal(err)
			}
			defer resp.Body.Close()

			var token string
			b, err := api.Unmarshal(resp.Body, api.KV{Key: "token", Value: &token})
			if err != nil {
				output.Fatal(err)
			}

			switch resp.StatusCode {
			case http.StatusOK:
				fmt.Printf("New token: %s\n", token)
			case http.StatusUnauthorized:
				fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, b.Message)
			default:
				fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
					resp.StatusCode, b.Message)
			}
		},
	}
	root.AddCommand(token)
}

func (root *HostCmd) attachUpgradeCmd() {
	const flagVersion = "version"
	var upgrade = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Inertia daemon to match the CLI.",
		Long:  `Restarts the Inertia daemon to upgrade it to the same version as your CLI`,
		Run: func(cmd *cobra.Command, args []string) {
			sshc, err := root.client.GetSSHClient()
			if err != nil {
				output.Fatal(err)
			}

			println("Shutting down daemon...")
			if err = sshc.DaemonDown(); err != nil {
				output.Fatal(err)
			}

			var version = root.getRemote().Version
			if v, _ := cmd.Flags().GetString(flagVersion); v != "" {
				version = v
			}

			fmt.Printf("Starting up the Inertia daemon (version %s)\n", version)
			if err := sshc.DaemonUp(); err != nil {
				output.Fatal(err)
			}
		},
	}
	upgrade.Flags().String(flagVersion, "", "version of Inertia daemon to spin up")
	root.AddCommand(upgrade)
}
