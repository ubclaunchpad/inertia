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

// Initialize "inertia [REMOTE] [COMMAND]" commands
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
			Use:    remote.Name + " [COMMAND]",
			Hidden: true,
			Short:  "Configure deployment to " + remote.Name,
			Long: `Manage deployment on specified remote.

Requires:
1. an Inertia daemon running on your remote - use 'inertia [REMOTE] init'
   to set one up.
2. a deploy key to be registered for the daemon with your GitHub repository.

Continuous deployment requires a webhook url to registered for the daemon
with your GitHub repository.

Run 'inertia [REMOTE] init' to gather this information.`,
		}

		// Deep copy and attach each deployment command.
		up := deepCopy(cmdDeploymentUp)
		up.Flags().String("type", "", "Specify a build method for your project")
		cmd.AddCommand(up)

		down := deepCopy(cmdDeploymentDown)
		cmd.AddCommand(down)

		status := deepCopy(cmdDeploymentStatus)
		cmd.AddCommand(status)

		logs := deepCopy(cmdDeploymentLogs)
		cmd.AddCommand(logs)

		user := deepCopy(cmdDeploymentUser)
		adduser := deepCopy(cmdDeploymentAddUser)
		adduser.Flags().Bool("admin", false, "Create an admin user")
		removeuser := deepCopy(cmdDeploymentRemoveUser)
		resetusers := deepCopy(cmdDeploymentResetUsers)
		listusers := deepCopy(cmdDeploymentListUsers)
		user.AddCommand(adduser)
		user.AddCommand(removeuser)
		user.AddCommand(resetusers)
		user.AddCommand(listusers)
		cmd.AddCommand(user)

		ssh := deepCopy(cmdDeploymentSSH)
		cmd.AddCommand(ssh)

		send := deepCopy(cmdDeploymentSendFile)
		send.Flags().StringP("dest", "d", "", "Path relative from project root to send file to")
		send.Flags().StringP("permissions", "p", "0655", "Permissions settings for file")
		cmd.AddCommand(send)

		env := deepCopy(cmdDeploymentEnv)
		setenv := deepCopy(cmdDeploymentEnvSet)
		setenv.Flags().BoolP("encrypt", "e", false, "Encrypt variable when stored")
		env.AddCommand(setenv)
		env.AddCommand(deepCopy(cmdDeploymentEnvRemove))
		env.AddCommand(deepCopy(cmdDeploymentEnvList))
		cmd.AddCommand(env)

		init := deepCopy(cmdDeploymentInit)
		cmd.AddCommand(init)

		reset := deepCopy(cmdDeploymentReset)
		cmd.AddCommand(reset)

		// Attach a "short" option on all commands
		cmd.PersistentFlags().BoolP(
			"short", "s", false,
			"Don't stream output from command",
		)
		// Attach "secure" option on all commands to enable SSL verification
		cmd.PersistentFlags().Bool(
			"verify-ssl", false,
			"Verify SSL communications - requires a signed SSL certificate.",
		)
		Root.AddCommand(cmd)
	}
}

var cmdDeploymentUp = &cobra.Command{
	Use:   "up",
	Short: "Bring project online on remote",
	Long: `Bring project online on remote.
	This will run 'docker-compose up --build'. Requires the Inertia daemon
	to be active on your remote - do this by running 'inertia [REMOTE] init'`,
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
	Long: `Bring project offline on remote.
	This will kill all active project containers on your remote.
	Requires project to be online - do this by running 'inertia [REMOTE] up`,
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
	Short: "Print the status of deployment on remote",
	Long: `Print the status of deployment on remote.
	Requires the Inertia daemon to be active on your remote - do this by 
	running 'inertia [REMOTE] up'`,
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
	Use:   "logs",
	Short: "Access logs of your VPS",
	Long: `Access logs of containers of your VPS. Argument 'docker-compose'
	will retrieve logs of the docker-compose build. The additional argument can 
	also be used to access logs of specific containers - use  'inertia [REMOTE] 
	status' to see what containers are accessible.`,
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

var cmdDeploymentSSH = &cobra.Command{
	Use:   "ssh",
	Short: "Start an interactive SSH session",
	Long:  `Starts up an interact SSH session with your remote.`,
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
	Use:   "send",
	Short: "Send a file to your Inertia deployment",
	Long: `Send a file, such as a configuration or .env file, to your Inertia
deployment. Provide a relative path to your file.`,
	Args: cobra.MinimumNArgs(1),
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
	Short: "Initialize the VPS for continuous deployment",
	Long: `Initialize the VPS for continuous deployment.
This sets up everything you might need and brings the Inertia daemon
online on your remote.
A URL will be provided to direct GitHub webhooks to, the daemon will
request access to the repository via a public key, and will listen
for updates to this repository's remote master branch.`,
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
	Long: `Reset the project on your remote.
On this remote, this kills all active containers and clears the project
directory, allowing you to assign a different Inertia project to this
remote. Requires Inertia daemon to be active on your remote - do this by
running 'inertia [REMOTE] init'`,
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
