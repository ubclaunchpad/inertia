package remotescmd

import (
	"fmt"
	"net/http"
	"syscall"

	qr "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/cmd/printutil"
)

// UserTotpCmd is the parent class for the 'user totp' subcommands
type UserTotpCmd struct {
	*cobra.Command
	host *HostCmd
}

// AttachTotpCmd attaches the 'totp' subcommands to given parent
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

			var totpInfo api.TotpResponse
			b, err := api.Unmarshal(resp.Body, api.KV{Key: "totp", Value: &totpInfo})
			if err != nil {
				printutil.Fatal(err)
			}

			// Display QR code so users can easily add their keys to their
			// authenticator apps
			qr.New().Get(fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=Inertia",
				username, totpInfo.TotpSecret)).Print()

			fmt.Printf("\n\n(Status code %d) %s\n",
				resp.StatusCode, b.Message)
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
