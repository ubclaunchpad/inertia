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
		resp, err := deployment.Up(stream)
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
			fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
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

// deploymentCmd represents the deploy command
var deploymentCmd = &cobra.Command{
	Hidden: true,
	Long: `Start or stop continuous deployment to the remote VPS instance specified.
Run 'inertia remote status' beforehand to ensure your daemon is running.
Requires:

1. A deploy key to be registered for the daemon with your GitHub repository.
2. A webhook url to registered for the daemon with your GitHub repository.

Run 'inertia [REMOTE] init' to collect these.`,
}

func init() {
	config, err := client.GetProjectConfigFromDisk()
	if err != nil {
		return
	}

	for _, remote := range config.Remotes {
		newCmd := &cobra.Command{}
		*newCmd = *deploymentCmd
		addDeploymentCommand(remote.Name, newCmd)
	}
}

func addDeploymentCommand(remoteName string, cmd *cobra.Command) {
	cmd.Use = remoteName + " [COMMAND]"
	cmd.Short = "Configure continuous deployment to " + remoteName
	cmd.AddCommand(deploymentUpCmd)
	cmd.AddCommand(deploymentDownCmd)
	cmd.AddCommand(deploymentStatusCmd)
	cmd.AddCommand(deploymentResetCmd)
	cmd.AddCommand(deploymentInitCmd)
	cmd.AddCommand(deploymentLogsCmd)
	rootCmd.AddCommand(cmd)

	cmd.PersistentFlags().Bool("stream", false, "Stream output from daemon")
}
