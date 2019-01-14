package hostcmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/ubclaunchpad/inertia/client"
	inertiacmd "github.com/ubclaunchpad/inertia/cmd/cmd"
	"github.com/ubclaunchpad/inertia/cmd/printutil"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"

	"github.com/spf13/cobra"
)

const (
	// EnvSSHPassphrase is the key used to fetch PEM key passphrases
	EnvSSHPassphrase = "PEM_PASSPHRASE"
)

// AttachHostCmds reads configuration to attach a child command for each
// configured remote in the configuration
func AttachHostCmds(inertia *inertiacmd.Cmd) {
	config, path, err := local.GetProjectConfigFromDisk(inertia.ConfigPath)
	if err != nil {
		return
	}
	if config.Version != inertia.Version {
		fmt.Printf("[WARNING] Configuration version '%s' does not match your Inertia CLI version '%s'\n",
			config.Version, inertia.Version)
	}
	for remote := range config.Remotes {
		attachHostCmd(inertia, remote, config, path)
	}
}

// HostCmd is the parent class for a subcommand for a configured remote host
type HostCmd struct {
	*cobra.Command
	remote  string
	config  *cfg.Config
	cfgPath string
	client  *client.Client
}

const (
	flagShort     = "short"
	flagVerifySSL = "verify-ssl"
)

// attachHostCmd attaches a subcommand for a configured remote host to the
// given parent
func attachHostCmd(inertia *inertiacmd.Cmd, remote string, config *cfg.Config, cfgPath string) {
	cli, found := client.NewClient(remote, os.Getenv(EnvSSHPassphrase), config, os.Stdout)
	if !found {
		printutil.Fatal("Remote not found")
	}
	var host = &HostCmd{
		remote:  remote,
		config:  config,
		cfgPath: cfgPath,
		client:  cli,
	}
	host.Command = &cobra.Command{
		Use:    remote + " [command]",
		Hidden: true,
		Short:  "Configure deployment to " + remote,
		Long: `Manages deployment on specified remote.

Requires:
1. an Inertia daemon running on your remote - use 'inertia [remote] init' to get it running.
2. a deploy key to be registered within your remote repository for the daemon to use.

Continuous deployment requires the daemon's webhook address to be registered in your remote repository.

If the SSH key for your remote requires a passphrase, it can be provided via 'PEM_PASSPHRASE'.

Run 'inertia [remote] init' to gather this information.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if host.client == nil || host.config == nil {
				printutil.Fatalf("failed to read configuration at '%s'", host.cfgPath)
			}
			var verify, _ = cmd.Flags().GetBool(flagVerifySSL)
			host.client.SetSSLVerification(verify)
		},
	}
	host.PersistentFlags().BoolP(flagShort, "s", false,
		"don't stream output from command")
	host.PersistentFlags().Bool(flagVerifySSL, false,
		"verify SSL communications - requires a signed SSL certificate")

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
	host.attachTokenCmd()
	host.attachUpgradeCmd()
	host.attachUninstallCmd()

	// attach to parent
	inertia.AddCommand(host.Command)
}

func (root *HostCmd) attachUpCmd() {
	const flagBuildType = "type"
	var up = &cobra.Command{
		Use:   "up",
		Short: "Bring project online on remote",
		Long: `Builds and deploy your project on your remote.

This requires an Inertia daemon to be active on your remote - do this by running 'inertia [remote] init'`,
		Run: func(cmd *cobra.Command, args []string) {
			// Get flags
			var short, _ = cmd.Flags().GetBool(flagShort)
			var buildType, _ = cmd.Flags().GetString(flagBuildType)

			// TODO: support other remotes
			url, err := local.GetRepoRemote("origin")
			if err != nil {
				printutil.Fatal(err)
			}

			resp, err := root.client.Up(url, buildType, !short)
			if err != nil {
				printutil.Fatal(err)
			}
			defer resp.Body.Close()

			if short {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					printutil.Fatal(err)
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
	up.Flags().String(flagBuildType, "", "override configured build method for your project")
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
				printutil.Fatal(err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				printutil.Fatal(err)
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
				printutil.Fatal(err)
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK:
				var host = "https://" + root.client.RemoteVPS.GetIPAndPort()
				fmt.Printf("(Status code %d) Daemon at remote '%s' online at %s\n",
					resp.StatusCode, root.client.Name, host)
				var status = &api.DeploymentStatus{}
				if err := json.NewDecoder(resp.Body).Decode(status); err != nil {
					printutil.Fatal(err)
				}
				println(printutil.FormatStatus(status))
			case http.StatusUnauthorized:
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					printutil.Fatal(err)
				}
				fmt.Printf("(Status code %d) Bad auth: %s\n", resp.StatusCode, body)
			default:
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					printutil.Fatal(err)
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
					printutil.Fatal(err)
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					printutil.Fatal(err)
				}
				switch resp.StatusCode {
				case http.StatusOK:
					fmt.Printf("(Status code %d) Logs: \n%s\n", resp.StatusCode, body)
				case http.StatusUnauthorized:
					fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
				case http.StatusPreconditionFailed:
					fmt.Printf("(Status code %d) Problem with deployment setup:\n%s\n", resp.StatusCode, body)
				default:
					fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
						resp.StatusCode, body)
				}
			} else {
				// if not short, open a websocket to stream logs
				socket, err := root.client.LogsWebSocket(container, entries)
				if err != nil {
					printutil.Fatal(err)
				}
				defer socket.Close()

				for {
					_, line, err := socket.ReadMessage()
					if err != nil {
						printutil.Fatal(err)
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
				printutil.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				printutil.Fatal(err)
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
			if err := root.client.SSH.RunSession(args...); err != nil {
				printutil.Fatal(err.Error())
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
				printutil.Fatal(err.Error())
			}
			f, err := os.Open(path.Join(cwd, args[0]))
			if err != nil {
				printutil.Fatal(err.Error())
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
			if err = root.client.SSH.CopyFile(f, remotePath, permissions); err != nil {
				printutil.Fatal(err.Error())
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
			url, err := local.GetRepoRemote("origin")
			if err != nil {
				printutil.Fatal(err)
			}
			var repoName = common.ExtractRepository(common.GetSSHRemoteURL(url))
			if err = root.client.BootstrapRemote(repoName); err != nil {
				printutil.Fatal(err)
			}
			root.config.Write(root.cfgPath)
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
				printutil.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				printutil.Fatal(err)
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
				printutil.Fatal("aborting")
			}

			// Daemon down
			println("Stopping project...")
			if _, err = root.client.Down(); err != nil {
				printutil.Fatal(err)
			}
			println("Stopping daemon...")
			if err = root.client.DaemonDown(); err != nil {
				printutil.Fatal(err)
			}
			println("Removing Inertia directories...")
			if err = root.client.UninstallInertia(); err != nil {
				printutil.Fatal(err)
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
				printutil.Fatal(err)
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				printutil.Fatal(err)
			}

			switch resp.StatusCode {
			case http.StatusOK:
				fmt.Printf("New token: %s\n", string(body))
			case http.StatusUnauthorized:
				fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, string(body))
			default:
				fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
					resp.StatusCode, body)
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
			println("Shutting down daemon...")
			if err := root.client.DaemonDown(); err != nil {
				printutil.Fatal(err)
			}

			var version = root.config.Version
			if v, _ := cmd.Flags().GetString(flagVersion); v != "" {
				version = v
			}

			fmt.Printf("Starting up the Inertia daemon (version %s)\n", version)
			if err := root.client.DaemonUp(version); err != nil {
				printutil.Fatal(err)
			}
		},
	}
	upgrade.Flags().String(flagVersion, root.config.Version, "version of Inertia daemon to spin up")
	root.AddCommand(upgrade)
}
