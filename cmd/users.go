package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"syscall"

	"github.com/qianlnk/qrcode"
	"github.com/ubclaunchpad/inertia/common"

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
	Use:   "add [user]",
	Short: "Create a user with access to Inertia Web",
	Long: `Creates a user with access to the Inertia Web application.

This user will be able to log in and view or configure the
deployment from the web app.

Use the --admin flag to create an admin user.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
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
		case http.StatusUnauthorized:
			fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
		default:
			fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
				resp.StatusCode, body)
		}
	},
}

var cmdDeploymentRemoveUser = &cobra.Command{
	Use:   "rm [user]",
	Short: "Remove a user",
	Long: `Removes the given user from Inertia's user database.

This user will no longer be able to log in and view or configure the deployment
from the web app.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
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
		case http.StatusUnauthorized:
			fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
		default:
			fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
				resp.StatusCode, body)
		}
	},
}

var cmdDeploymentLogin = &cobra.Command{
	Use:   "login [user]",
	Short: "Authenticate with the remote",
	Long:  "Retreives an access token from the remote using your credentials.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		username := args[0]
		fmt.Print("Password: ")
		pwBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			log.Fatal(err)
		}

		resp, err := deployment.LogIn(username, string(pwBytes), "")
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode == http.StatusExpectationFailed {
			// a TOTP is required
			fmt.Print("TOTP: ")
			totpBytes, err := terminal.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				log.Fatal(err)
			}
			resp, err = deployment.LogIn(username, string(pwBytes), string(totpBytes))
			if err != nil {
				log.Fatal(err)
			}
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Invalid credentials")
			return
		}

		config, path, err := local.GetProjectConfigFromDisk(configFilePath)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()
		token, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		config.Remotes[remoteName].Daemon.Token = string(token)
		config.Remotes[remoteName].User = username
		err = config.Write(path)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("You have been logged in successfully.")
	},
}

var cmdDeploymentResetUsers = &cobra.Command{
	Use:   "reset",
	Short: "Reset user database on your remote",
	Long: `Removes all users credentials on your remote. All users will no longer
be able to log in and view or configure the deployment from the web app.`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
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
		case http.StatusUnauthorized:
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
	Long:  `Lists all users registered in Inertia's user database.`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
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
		case http.StatusUnauthorized:
			fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
		default:
			fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
				resp.StatusCode, body)
		}
	},
}

var cmdDeploymentEnableTotp = &cobra.Command{
	Use:   "enable-totp [user]",
	Short: "Enable Totp for a user",
	Long:  "Enable Totp for a user on your remote",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		username := args[0]
		fmt.Print("Password: ")
		pwBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			log.Fatal(err)
		}

		// Endpoint handles user authentication before enabling Totp
		resp, err := deployment.EnableTotp(username, string(pwBytes))
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("(Status code %d) Error Enabling Totp.\n", resp.StatusCode)
			return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var totpInfo common.TotpResponse
		err = json.Unmarshal(body, &totpInfo)

		if err != nil {
			fmt.Println("Failed to Unmarshal Totp key")
			return
		}

		qr := qrcode.NewQRCode(totpInfo.TotpSecret, true)
		qr.Output()

		fmt.Printf("Totp successfully enabled. "+
			"Your secret key is %s and your backup codes are %v\n",
			totpInfo.TotpSecret, totpInfo.BackupCodes)
	},
}

var cmdDeploymentDisableTotp = &cobra.Command{
	Use:   "disable-totp [user]",
	Short: "Disable Totp for a user",
	Long:  "Disable Totp for a user on your remote",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		username := args[0]
		fmt.Print("Password: ")
		pwBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print("TOTP: ")
		totpBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			log.Fatal(err)
		}

		// Endpoint handles user authentication before disabling Totp
		resp, err := deployment.DisableTotp(username, string(pwBytes), string(totpBytes))
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Error Disabling Totp. Status Code: " + string(resp.StatusCode))
			return
		}

		fmt.Println("Totp successfully disabled.")
	},
}
