package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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
		// Start the deployment
		deployment, err := client.GetDeployment(strings.Split(cmd.Parent().Use, " ")[0])
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
				log.WithError(err)
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
		// Shut down the deployment
		deployment, err := client.GetDeployment(strings.Split(cmd.Parent().Use, " ")[0])
		if err != nil {
			log.Fatal(err)
		}
		resp, err := deployment.Down()
		if err != nil {
			log.WithError(err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithError(err)
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

		// Get status of the deployment
		deployment, err := client.GetDeployment(strings.Split(cmd.Parent().Use, " ")[0])
		if err != nil {
			log.Fatal(err)
		}
		host := "http://" + deployment.RemoteVPS.GetIPAndPort()
		resp, err := deployment.Status()
		if err != nil {
			log.WithError(err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithError(err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			fmt.Printf("(Status code %d) Daemon at remote '%s' online at %s\n", resp.StatusCode, deployment.Name, host)
			fmt.Printf("%s", body)
		case http.StatusForbidden:
			fmt.Printf("(Status code %d) Bad auth: %s\n", resp.StatusCode, body)
		case http.StatusNotFound:
			fmt.Printf("(Status code %d) Problem with deployment: %s\n", resp.StatusCode, body)
		case http.StatusPreconditionFailed:
			fmt.Printf("(Status code %d) Problem with deployment setup: %s\n", resp.StatusCode, body)
		default:
			fmt.Printf("(Status code %d) Unknown response from daemon: %s\n",
				resp.StatusCode, body)
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
		// Remove project from deployment
		deployment, err := client.GetDeployment(strings.Split(cmd.Parent().Use, " ")[0])
		if err != nil {
			log.Fatal(err)
		}
		resp, err := deployment.Reset()
		if err != nil {
			log.WithError(err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithError(err)
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

var deploymentLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Access logs of your VPS",
	Long: `Access logs of containers of your VPS. Argument 'docker-compose'
	will retrieve logs of the docker-compose build. The additional argument can 
	also be used to access logs of specific containers - use  'inertia [REMOTE] 
	status' to see what containers are accessible.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Start the deployment
		deployment, err := client.GetDeployment(strings.Split(cmd.Parent().Use, " ")[0])
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
				log.WithError(err)
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

// deploymentSSHCmd represents the inertia [REMOTE] ssh command
var deploymentSSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Start an interactive SSH session",
	Long:  `Starts up an interact SSH session with your remote.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
		remote, found := config.GetRemote(remoteName)
		if found {
			session := client.NewSSHRunner(remote)
			if err = session.RunSession(); err != nil {
				log.Fatal(err.Error())
			}
		} else {
			log.Fatal(errors.New("There does not appear to be a remote with this name. Have you modified the Inertia configuration file?"))
		}
	},
}

// deploymentInitCmd represents the inertia [REMOTE] init command
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
		// Ensure project initialized.
		config, err := client.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		remoteName := strings.Split(cmd.Parent().Use, " ")[0]
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

func init() {
	config, err := client.GetProjectConfigFromDisk()
	if err != nil {
		return
	}

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

		up := &cobra.Command{}
		*up = *deploymentUpCmd
		up.Flags().String("type", "", "Specify a build method for your project")
		cmd.AddCommand(up)

		down := &cobra.Command{}
		*down = *deploymentDownCmd
		cmd.AddCommand(down)

		status := &cobra.Command{}
		*status = *deploymentStatusCmd
		cmd.AddCommand(status)

		reset := &cobra.Command{}
		*reset = *deploymentResetCmd
		cmd.AddCommand(reset)

		logs := &cobra.Command{}
		*logs = *deploymentLogsCmd
		cmd.AddCommand(logs)

		ssh := &cobra.Command{}
		*ssh = *deploymentSSHCmd
		cmd.AddCommand(ssh)

		init := &cobra.Command{}
		*init = *deploymentInitCmd
		cmd.AddCommand(init)

		cmd.PersistentFlags().Bool("stream", false, "Stream output from daemon")
		rootCmd.AddCommand(cmd)
	}
}
