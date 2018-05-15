package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ubclaunchpad/inertia/common"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/client"
)

var deploymentUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Bring project online on remote",
	Long: `Bring project online on remote.
	This will run 'docker-compose up --build'. Requires the Inertia daemon
	to be active on your remote - do this by running 'inertia [REMOTE] init'`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, err := client.GetDeployment(remoteName)
		if err != nil {
			log.Fatal(err)
		}
		stream, err := cmd.Flags().GetBool("stream")

		if err != nil {
			log.Fatal(err)
		}
		buildType, err := cmd.Flags().GetString("type")
		if err != nil {
			log.Fatal(err)
		}
		resp, err := deployment.Up(buildType, stream)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if !stream {
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

var deploymentDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Bring project offline on remote",
	Long: `Bring project offline on remote.
	This will kill all active project containers on your remote.
	Requires project to be online - do this by running 'inertia [REMOTE] up`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, err := client.GetDeployment(remoteName)
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

var deploymentStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the status of deployment on remote",
	Long: `Print the status of deployment on remote.
	Requires the Inertia daemon to be active on your remote - do this by 
	running 'inertia [REMOTE] up'`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, err := client.GetDeployment(remoteName)
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

var deploymentLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Access logs of your VPS",
	Long: `Access logs of containers of your VPS. Argument 'docker-compose'
	will retrieve logs of the docker-compose build. The additional argument can 
	also be used to access logs of specific containers - use  'inertia [REMOTE] 
	status' to see what containers are accessible.`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, err := client.GetDeployment(remoteName)
		if err != nil {
			log.Fatal(err)
		}
		stream, err := cmd.Flags().GetBool("stream")
		if err != nil {
			log.Fatal(err)
		}

		container := "/inertia-daemon"
		if len(args) > 0 {
			container = args[0]
		}

		resp, err := deployment.Logs(stream, container)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if !stream {
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

var deploymentSSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Start an interactive SSH session",
	Long:  `Starts up an interact SSH session with your remote.`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, err := client.GetDeployment(remoteName)
		if err != nil {
			log.Fatal(err)
		}

		session := client.NewSSHRunner(deployment.RemoteVPS)
		if err = session.RunSession(); err != nil {
			log.Fatal(err.Error())
		}
	},
}

var deploymentInitCmd = &cobra.Command{
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

		// Bootstrap needs to write to configuration.
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}
		remote, found := config.GetRemote(remoteName)
		if found {
			session := client.NewSSHRunner(remote)
			err = remote.Bootstrap(session, remoteName, config)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(errors.New("There does not appear to be a remote with this name. Have you modified the Inertia configuration file?"))
		}
	},
}

var deploymentResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the project on your remote",
	Long: `Reset the project on your remote.
On this remote, this kills all active containers and clears the project
directory, allowing you to assign a different Inertia project to this
remote. Requires Inertia daemon to be active on your remote - do this by
running 'inertia [REMOTE] init'`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		deployment, err := client.GetDeployment(remoteName)
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

func init() {
	config, err := client.GetProjectConfigFromDisk()
	if err != nil {
		return
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
		up := deepCopy(deploymentUpCmd)
		up.Flags().String("type", "", "Specify a build method for your project")
		cmd.AddCommand(up)

		down := deepCopy(deploymentDownCmd)
		cmd.AddCommand(down)

		status := deepCopy(deploymentStatusCmd)
		cmd.AddCommand(status)

		logs := deepCopy(deploymentLogsCmd)
		cmd.AddCommand(logs)

		user := deepCopy(deploymentUserCmd)
		adduser := deepCopy(deploymentUserAddCmd)
		adduser.Flags().Bool("admin", false, "Create an admin user")
		removeuser := deepCopy(deploymentUserRemoveCmd)
		resetusers := deepCopy(deploymentUsersResetCmd)
		listusers := deepCopy(deploymentUsersListCmd)
		user.AddCommand(adduser)
		user.AddCommand(removeuser)
		user.AddCommand(resetusers)
		user.AddCommand(listusers)
		cmd.AddCommand(user)

		ssh := deepCopy(deploymentSSHCmd)
		cmd.AddCommand(ssh)

		init := deepCopy(deploymentInitCmd)
		cmd.AddCommand(init)

		reset := deepCopy(deploymentResetCmd)
		cmd.AddCommand(reset)

		// Attach a "stream" option on all commands, even if it doesn't
		// do anything for some commands yet.
		cmd.PersistentFlags().BoolP(
			"stream", "s", false,
			"Stream output from daemon - doesn't do anything on some commands.",
		)
		rootCmd.AddCommand(cmd)
	}
}

// deepCopy is a helper function for deeply copying a command.
func deepCopy(cmd *cobra.Command) *cobra.Command {
	newCmd := &cobra.Command{}
	*newCmd = *cmd
	return newCmd
}
