package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/client"
)

// parseConfigArg is a dirty dirty hack to allow access to the --config argument
// before Cobra parses it (it is required to set up remote commands in the
// init() phase)
func parseConfigArg() {
	for i, arg := range os.Args {
		if arg == "--config" {
			configFilePath = os.Args[i+1]
			break
		}
	}
}

// Initialize "inertia [remote] [cmd]" commands
func init() {
	// This is the only place configuration is read every time an `inertia`
	// command is run - check version here.
	parseConfigArg() // see parseArgs documentation
	config, _, err := local.GetProjectConfigFromDisk(configFilePath)
	if err != nil {
		println("[WARNING] Inertia configuration not found in " + configFilePath)
		return
	}
	if config.Version != Root.Version {
		fmt.Printf(
			"[WARNING] Configuration version '%s' does not match your Inertia CLI version '%s'\n",
			config.Version, Root.Version,
		)
	}

	// Make a new command for each remote with all associated
	// deployment commands.
	for _, remote := range config.Remotes {
		cmd := &cobra.Command{
			Use:    remote.Name + " [command]",
			Hidden: true,
			Short:  "Configure deployment to " + remote.Name,
			Long: `Manages deployment on specified remote.

Requires:
1. an Inertia daemon running on your remote - use 'inertia [remote] init' to get it running.
2. a deploy key to be registered within your remote repository for the daemon to use.

Continuous deployment requires the daemon's webhook address to be registered in your remote repository.

Run 'inertia [remote] init' to gather this information.`,
		}

		// Deep copy and attach each deployment command.
		up := deepCopy(cmdDeploymentUp)
		up.Flags().String("type", "", "specify a build method for your project")
		cmd.AddCommand(up)

		cmd.AddCommand(deepCopy(cmdDeploymentDown))
		cmd.AddCommand(deepCopy(cmdDeploymentStatus))
		cmd.AddCommand(deepCopy(cmdDeploymentLogs))
		cmd.AddCommand(deepCopy(cmdDeploymentPrune))

		user := deepCopy(cmdDeploymentUser)
		adduser := deepCopy(cmdDeploymentAddUser)
		adduser.Flags().Bool("admin", false, "create a user with administrator permissions")
		user.AddCommand(adduser)
		user.AddCommand(deepCopy(cmdDeploymentRemoveUser))
		user.AddCommand(deepCopy(cmdDeploymentResetUsers))
		user.AddCommand(deepCopy(cmdDeploymentListUsers))
		cmd.AddCommand(user)

		cmd.AddCommand(deepCopy(cmdDeploymentSSH))

		send := deepCopy(cmdDeploymentSendFile)
		send.Flags().StringP("dest", "d", "", "path relative from project root to send file to")
		send.Flags().StringP("permissions", "p", "0655", "permissions settings to create file with")
		cmd.AddCommand(send)

		env := deepCopy(cmdDeploymentEnv)
		setenv := deepCopy(cmdDeploymentEnvSet)
		setenv.Flags().BoolP("encrypt", "e", false, "encrypt variable when stored")
		env.AddCommand(setenv)
		env.AddCommand(deepCopy(cmdDeploymentEnvRemove))
		env.AddCommand(deepCopy(cmdDeploymentEnvList))
		cmd.AddCommand(env)

		cmd.AddCommand(deepCopy(cmdDeploymentInit))
		cmd.AddCommand(deepCopy(cmdDeploymentReset))

		remove := deepCopy(cmdDeploymentRemove)
		cmd.AddCommand(remove)

		// Attach a "short" option on all commands
		cmd.PersistentFlags().BoolP(
			"short", "s", false,
			"don't stream output from command",
		)
		// Attach "secure" option on all commands to enable SSL verification
		cmd.PersistentFlags().Bool(
			"verify-ssl", false,
			"verify SSL communications - requires a signed SSL certificate",
		)
		Root.AddCommand(cmd)
	}
}

var cmdDeploymentUp = &cobra.Command{
	Use:   "up",
	Short: "Bring project online on remote",
	Long: `Builds and deploy your project on your remote.

This requires an Inertia daemon to be active on your remote - do this by running 'inertia [remote] init'`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		// Get flags
		short, err := cmd.Flags().GetBool("short")
		if err != nil {
			log.Fatal(err)
		}
		buildType, err := cmd.Flags().GetString("type")
		if err != nil {
			log.Fatal(err)
		}

		// TODO: support other remotes
		url, err := local.GetRepoRemote("origin")
		if err != nil {
			log.Fatal(err)
		}

		resp, err := deployment.Up(url, buildType, !short)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if short {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			switch resp.StatusCode {
			case http.StatusCreated:
				fmt.Printf("(Status code %d) Project build started!\n", resp.StatusCode)
			case http.StatusForbidden:
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

var cmdDeploymentDown = &cobra.Command{
	Use:   "down",
	Short: "Bring project offline on remote",
	Long: `Stops your project on your remote. This will kill all active project containers on your remote.

Requires project to be online - do this by running 'inertia [remote] up`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := deployment.Down()
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			fmt.Printf("(Status code %d) Project down\n", resp.StatusCode)
		case http.StatusPreconditionFailed:
			fmt.Printf("(Status code %d) No containers are currently active\n", resp.StatusCode)
		case http.StatusForbidden:
			fmt.Printf("(Status code %d) Bad auth: %s\n", resp.StatusCode, body)
		default:
			fmt.Printf("(Status code %d) Unknown response from daemon: %s\n",
				resp.StatusCode, body)
		}
	},
}

var cmdDeploymentStatus = &cobra.Command{
	Use:   "status",
	Short: "Print the status of the deployment on this remote",
	Long: `Prints the status of the deployment on this remote.
Requires the Inertia daemon to be active on your remote - do this by running 'inertia [remote] up'`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}
		host := "http://" + deployment.RemoteVPS.GetIPAndPort()
		resp, err := deployment.Status()
		if err != nil {
			log.Fatal(err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			fmt.Printf("(Status code %d) Daemon at remote '%s' online at %s\n", resp.StatusCode, deployment.Name, host)
			status := &common.DeploymentStatus{}
			err := json.NewDecoder(resp.Body).Decode(status)
			if err != nil {
				log.Fatal(err)
			}
			println(formatStatus(status))
		case http.StatusForbidden:
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			fmt.Printf("(Status code %d) Bad auth: %s\n", resp.StatusCode, body)
		default:
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			fmt.Printf("(Status code %d) %s\n",
				resp.StatusCode, body)
		}
	},
}

var cmdDeploymentLogs = &cobra.Command{
	Use:   "logs [container]",
	Short: "Access logs of containers on your remote host",
	Long: `Accesses logs of containers on your remote host. 
	
By default, this command retrieves Inertia daemon logs, but you can provide an 
argument that specifies the name of the container you wish to retrieve logs for. 
Use 'inertia [remote] status' to see which containers are active.`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}
		short, err := cmd.Flags().GetBool("short")
		if err != nil {
			log.Fatal(err)
		}

		container := "/inertia-daemon"
		if len(args) > 0 {
			container = args[0]
		}

		if short {
			resp, err := deployment.Logs(container)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			switch resp.StatusCode {
			case http.StatusOK:
				fmt.Printf("(Status code %d) Logs: \n%s\n", resp.StatusCode, body)
			case http.StatusForbidden:
				fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
			case http.StatusPreconditionFailed:
				fmt.Printf("(Status code %d) Problem with deployment setup:\n%s\n", resp.StatusCode, body)
			default:
				fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
					resp.StatusCode, body)
			}
		} else {
			socket, err := deployment.LogsWebSocket(container)
			if err != nil {
				log.Fatal(err)
			}
			defer socket.Close()

			for {
				_, line, err := socket.ReadMessage()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Print(string(line))
			}
		}
	},
}

var cmdDeploymentPrune = &cobra.Command{
	Use:   "prune",
	Short: "Prune Docker assets and images on your remote",
	Long:  `Prunes Docker assets and images from your remote to free up storage space.`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		inertia, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := inertia.Prune()
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
	},
}

var cmdDeploymentSSH = &cobra.Command{
	Use:   "ssh",
	Short: "Start an interactive SSH session",
	Long:  `Starts an interact SSH session with your remote.`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath)
		if err != nil {
			log.Fatal(err)
		}

		session := client.NewSSHRunner(deployment.RemoteVPS)
		if err = session.RunSession(); err != nil {
			log.Fatal(err.Error())
		}
	},
}

var cmdDeploymentSendFile = &cobra.Command{
	Use:   "send [filepath]",
	Short: "Send a file to your Inertia deployment",
	Long:  `Sends a file, such as a configuration or .env file, to your Inertia deployment.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		// Get permissions to copy file with
		permissions, err := cmd.Flags().GetString("permissions")
		if err != nil {
			log.Fatal(err.Error())
		}

		// Open file with given name
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err.Error())
		}
		f, err := os.Open(path.Join(cwd, args[0]))
		if err != nil {
			log.Fatal(err.Error())
		}

		// Get flag for destination
		dest, err := cmd.Flags().GetString("dest")
		if err != nil {
			log.Fatal(err.Error())
		}
		if dest == "" {
			dest = args[0]
		}

		// Destination path - todo: allow config
		projectPath := "$HOME/inertia/project"
		remotePath := path.Join(projectPath, dest)

		// Initiate copy
		session := client.NewSSHRunner(deployment.RemoteVPS)
		err = session.CopyFile(f, remotePath, permissions)
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println("File", args[0], "has been copied to", remotePath, "on remote", remoteName)
	},
}

var cmdDeploymentInit = &cobra.Command{
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
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		cli, write, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		url, err := local.GetRepoRemote("origin")
		if err != nil {
			log.Fatal(err)
		}
		repoName := common.ExtractRepository(common.GetSSHRemoteURL(url))
		err = cli.BootstrapRemote(repoName)
		if err != nil {
			log.Fatal(err)
		}
		write()
	},
}

var cmdDeploymentReset = &cobra.Command{
	Use:   "reset",
	Short: "Reset the project on your remote",
	Long: `Resets the project on your remote.
On this remote, this kills all active containers and clears the project directory, 
allowing you to assign a different Inertia project to this remote.`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := deployment.Reset()
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
		case http.StatusForbidden:
			fmt.Printf("(Status code %d) Bad auth: %s\n", resp.StatusCode, body)
		default:
			fmt.Printf("(Status code %d) Unknown response from daemon: %s\n",
				resp.StatusCode, body)
		}
	},
}

var cmdDeploymentRemove = &cobra.Command{
	Use:   "remove",
	Short: "Shut down Inertia and remove Inertia assets from remote host",
	Long: `Shuts down and removes the Inertia daemon, and removes the Inertia 
directory (~/inertia) from your remote host.`,
	Run: func(cmd *cobra.Command, args []string) {
		println("WARNING: This will remove Inertia from the remote")
		println("as well as take the daemon and is irreversible. Continue? (y/n)")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil || response != "y" {
			log.Fatal("aborting")
		}

		// Daemon down
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}
		println("Stopping daemon...")
		err = deployment.DaemonDown()
		if err != nil {
			log.Fatal(err)
		}
		println("Removing Inertia directories...")
		err = deployment.InertiaDown()
		if err != nil {
			log.Fatal(err)
		}

		println("Inertia and related daemon removed.")
	},
}
