package hostcmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"syscall"

	qr "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/cmd/printutil"
	"golang.org/x/crypto/ssh/terminal"
)

type UserCmd struct {
	*cobra.Command
	host *HostCmd
}

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
	var add = &cobra.Command{
		Use:   "add [user]",
		Short: "Create a user with access to Inertia Web",
		Long: `Creates a user with access to the Inertia Web application.
	
	This user will be able to log in and view or configure the
	deployment from the web app.
	
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

			var admin, _ = cmd.Flags().GetBool("admin")
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
	add.Flags().Bool("admin", false, "create a user with administrator permissions")
	root.AddCommand(add)
}

func (root *UserCmd) attachRemoveCmd() {
	var remove = &cobra.Command{
		Use:   "rm [user]",
		Short: "Remove a user",
		Long: `Removes the given user from Inertia's user database.
	
	This user will no longer be able to log in and view or configure the deployment
	from the web app.`,
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
			token, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				printutil.Fatal(err)
			}

			var config = root.host.config
			var remote = root.host.remote
			config.Remotes[remote].Daemon.Token = string(token)
			config.Remotes[remote].User = username
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
		Long: `Removes all users credentials on your remote. All users will no longer
	be able to log in and view or configure the deployment from the web app.`,
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

type UserTotpCmd struct {
	*cobra.Command
	host *HostCmd
}

func AttachTotpCmd(root *UserCmd) {
	var totp = &UserTotpCmd{
		Command: &cobra.Command{
			Use:   "totp",
			Short: "Manage TOTP settings for a user",
			Long:  "Manage TOTP settings for a registered user on your Inertia daemon",
		},
		host: root.host,
	}

	// attach children
	totp.attachEnableCmd()
	totp.attachDisableCmd()

	// attach to parent
	root.AddCommand(totp.Command)
}

func (root *UserTotpCmd) attachEnableCmd() {
	var enable = &cobra.Command{
		Use:   "enable [user]",
		Short: "Enable TOTP for a user",
		Long:  "Enable TOTP for a user on your remote",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var username = args[0]
			fmt.Print("Password: ")
			pwBytes, err := terminal.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				printutil.Fatal(err)
			}

			// Endpoint handles user authentication before enabling Totp
			resp, err := root.host.client.EnableTotp(username, string(pwBytes))
			if err != nil {
				printutil.Fatal(err)
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("(Status code %d) Error Enabling Totp.", resp.StatusCode)
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				printutil.Fatal(err)
			}

			var totpInfo api.TotpResponse
			if err = json.Unmarshal(body, &totpInfo); err != nil {
				printutil.Fatal(err)
			}

			// Display QR code so users can easily add their keys to their
			// authenticator apps
			qr.New().Get(fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=Inertia",
				username, totpInfo.TotpSecret)).Print()

			fmt.Printf("\n\n(Status code %d) TOTP successfully enabled.\n",
				resp.StatusCode)
			fmt.Print("Scan the QR code above to " +
				"add your Inertia account to your authenticator app.\n\n")
			fmt.Printf("Your secret key is: %s\n", totpInfo.TotpSecret)
			fmt.Print("Your backup codes are:\n\n")

			for _, backupCode := range totpInfo.BackupCodes {
				fmt.Println(backupCode)
			}

			fmt.Println("\nIMPORTANT: Store our backup codes somewhere safe. " +
				"If you lose your authentication device you will need to use them " +
				"to regain access to your account.")
		},
	}
	root.AddCommand(enable)
}

func (root *UserTotpCmd) attachDisableCmd() {
	var disable = &cobra.Command{
		Use:   "disable [user]",
		Short: "Disable TOTP for a user",
		Long:  "Disable TOTP for a user on your remote",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Endpoint handles user authentication before disabling Totp
			resp, err := root.host.client.DisableTotp()
			if err != nil {
				printutil.Fatal(err)
			}

			fmt.Printf("(Status code %d) ", resp.StatusCode)
			if resp.StatusCode == http.StatusUnauthorized {
				fmt.Println("Please try logging in again before " +
					"disabling two-factor authentication.")
			} else if resp.StatusCode != http.StatusOK {
				fmt.Println("Error Disabling Totp.")
			} else {
				fmt.Println("Totp successfully disabled.")
			}
		},
	}
	root.AddCommand(disable)
}
