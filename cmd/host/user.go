package hostcmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/cmd/printutil"
	"golang.org/x/crypto/ssh/terminal"
)

// UserCmd is the parent class for the 'user' subcommands
type UserCmd struct {
	*cobra.Command
	host *HostCmd
}

// AttachUserCmd attaches the 'user' subcommands to the given parent
func AttachUserCmd(host *HostCmd) {
	var user = &UserCmd{
		Command: &cobra.Command{
			Use:   "user",
			Short: "Configure user access to Inertia Web",
			Long:  `Configure user access to the Inertia Web application.`,
		},
		host: host,
	}

	// attach children
	user.attachLoginCmd()
	AttachTotpCmd(user)
	user.attachAddCmd()
	user.attachRemoveCmd()
	user.attachListCmd()
	user.attachResetCmd()

	// attach to parent
	host.AddCommand(user.Command)
}

func (root *UserCmd) attachAddCmd() {
	const flagAdmin = "admin"
	var add = &cobra.Command{
		Use:   "add [user]",
		Short: "Create a user with access to this remote's Inertia daemon",
		Long: `Creates a user with access to this remote's Inertia daemon.

This user will be able to log in and view or configure the deployment
from the Inertia CLI (using 'inertia [remote] user login').

Use the --admin flag to create an admin user.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print("Enter a password for user: ")
			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				printutil.Fatal("Invalid password")
			}
			var password = strings.TrimSpace(string(bytePassword))
			fmt.Print("\n")

			var admin, _ = cmd.Flags().GetBool(flagAdmin)
			resp, err := root.host.client.AddUser(args[0], password, admin)
			if err != nil {
				printutil.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				printutil.Fatal(err)
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
	add.Flags().Bool(flagAdmin, false, "create a user with administrator permissions")
	root.AddCommand(add)
}

func (root *UserCmd) attachRemoveCmd() {
	var remove = &cobra.Command{
		Use:   "rm [user]",
		Short: "Remove a user",
		Long: `Removes the given user from Inertia's user database.

This user will no longer be able to log in and view or configure the deployment
remotely.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := root.host.client.RemoveUser(args[0])
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
				fmt.Printf("(Status code %d) User removed.\n", resp.StatusCode)
			case http.StatusUnauthorized:
				fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
			default:
				fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
					resp.StatusCode, body)
			}
		},
	}
	root.AddCommand(remove)
}

func (root *UserCmd) attachLoginCmd() {
	var login = &cobra.Command{
		Use:   "login [user]",
		Short: "Authenticate with the remote",
		Long:  "Retreives an access token from the remote using your credentials.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var username = args[0]
			fmt.Print("Password: ")
			pwBytes, err := terminal.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				printutil.Fatal(err)
			}

			var totp, _ = cmd.Flags().GetString("totp")
			resp, err := root.host.client.LogIn(username, string(pwBytes), totp)
			if err != nil {
				printutil.Fatal(err)
			}

			if resp.StatusCode == http.StatusExpectationFailed {
				// a TOTP is required
				fmt.Print("Authentication code (or backup code): ")
				totpBytes, err := terminal.ReadPassword(int(syscall.Stdin))
				fmt.Println()
				if err != nil {
					printutil.Fatal(err)
				}
				resp, err = root.host.client.LogIn(username, string(pwBytes), string(totpBytes))
				if err != nil {
					printutil.Fatal(err)
				}
			}

			fmt.Printf("(Status code %d) ", resp.StatusCode)
			if resp.StatusCode != http.StatusOK {
				fmt.Println("Invalid credentials")
				return
			}
			defer resp.Body.Close()
			var token string
			if api.Unmarshal(resp.Body, api.KV{Key: "token", Value: &token}); err != nil {
				printutil.Fatal(err)
			}

			var config = root.host.config
			var remote = root.host.remote
			config.Remotes[remote].Daemon.Token = string(token)
			if err = config.Write(root.host.cfgPath); err != nil {
				printutil.Fatal(err)
			}

			fmt.Println("You have been logged in successfully.")
		},
	}
	login.Flags().String("totp", "", "auth code or backup code for 2FA")
	root.AddCommand(login)
}

func (root *UserCmd) attachResetCmd() {
	var reset = &cobra.Command{
		Use:   "reset",
		Short: "Reset user database on your remote",
		Long: `Removes all users credentials on your remote. All configured user
will no longer be able to log in and view or configure the deployment
remotely.`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := root.host.client.ResetUsers()
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
				fmt.Printf("(Status code %d) All users removed.\n", resp.StatusCode)
			case http.StatusUnauthorized:
				fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
			default:
				fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
					resp.StatusCode, body)
			}
		},
	}
	root.AddCommand(reset)
}

func (root *UserCmd) attachListCmd() {
	var list = &cobra.Command{
		Use:   "ls",
		Short: "List all users registered on your remote.",
		Long:  `Lists all users registered in Inertia's user database.`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := root.host.client.ListUsers()
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
				fmt.Printf("(Status code %d) Bad auth:\n%s\n", resp.StatusCode, body)
			default:
				fmt.Printf("(Status code %d) Unknown response from daemon:\n%s\n",
					resp.StatusCode, body)
			}
		},
	}
	root.AddCommand(list)
}
