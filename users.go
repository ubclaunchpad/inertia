package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/local"
	"golang.org/x/crypto/ssh/terminal"
)

var cmdDeploymentUser = &cobra.Command{
	Use:   "user",
	Short: "Configure user access to Inertia Web",
	Long:  `Configure user access to the Inertia Web application.`,
}

var cmdDeploymentAddUser = &cobra.Command{
	Use:   "add",
	Short: "Create a user with access to Inertia Web",
	Long: `Create a user with access to the Inertia Web application.

This user will be able to log in and view or configure the
deployment from the web app.

Use the --admin flag to create an admin user.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, err := local.GetClient(remoteName, ConfigFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}
		admin, err := cmd.Flags().GetBool("admin")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print("Enter a password for user: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal("Invalid password")
		}
		password := strings.TrimSpace(string(bytePassword))
		fmt.Print("\n")

		resp, err := deployment.AddUser(args[0], password, admin)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		switch resp.StatusCode {
		case http.StatusCreated:
			fmt.Printf("(Status code %d) User added!\n", resp.StatusCode)
		case http.StatusForbidden:
			fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
		default:
			fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
				resp.StatusCode, body)
		}
	},
}

var cmdDeploymentRemoveUser = &cobra.Command{
	Use:   "rm",
	Short: "Remove a user with access to Inertia Web",
	Long: `Remove a user with access to the Inertia Web application.

This user will no longer be able to log in and view or configure the
deployment from the web app.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, err := local.GetClient(remoteName, ConfigFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := deployment.RemoveUser(args[0])
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithError(err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			fmt.Printf("(Status code %d) User removed.\n", resp.StatusCode)
		case http.StatusForbidden:
			fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
		default:
			fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
				resp.StatusCode, body)
		}
	},
}

var cmdDeploymentResetUsers = &cobra.Command{
	Use:   "reset",
	Short: "Reset user database on your remote.",
	Long: `Removes all users credentials on your remote. All users will
no longer be able to log in and view or configure the deployment 
from the web app.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, err := local.GetClient(remoteName, ConfigFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := deployment.ResetUsers()
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithError(err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			fmt.Printf("(Status code %d) All users removed.\n", resp.StatusCode)
		case http.StatusForbidden:
			fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
		default:
			fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
				resp.StatusCode, body)
		}
	},
}

var cmdDeploymentListUsers = &cobra.Command{
	Use:   "ls",
	Short: "List all users registered on your remote.",
	Long:  `List all users with access to Inertia Web on your remote.`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, err := local.GetClient(remoteName, ConfigFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := deployment.ListUsers()
		if err != nil {
			log.Fatal(err)
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
			fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
		default:
			fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
				resp.StatusCode, body)
		}
	},
}
