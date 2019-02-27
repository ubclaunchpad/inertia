package remotescmd

import (
	"context"
	"fmt"
	"syscall"

	qr "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/output"
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
			Use:     "totp",
			Short:   "Manage 2FA TOTP settings for users",
			Long:    "Manage 2FA TOTP settings for registered users on your Inertia daemon",
			Aliases: []string{"2fa"},
		},
		host: root.host,
	}

	// attach children
	totp.attachEnableCmd()
	totp.attachDisableCmd()

	// attach to parent
	root.AddCommand(totp.Command)
}

// context returns the root host command's context
func (root *UserTotpCmd) context() context.Context { return root.host.ctx }

func (root *UserTotpCmd) getUserClient() *client.UserClient { return root.host.client.GetUserClient() }

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
				output.Fatal(err)
			}

			// Endpoint handles user authentication before enabling Totp
			totpInfo, err := root.getUserClient().EnableTotp(root.context(), username, string(pwBytes))
			if err != nil {
				output.Fatal(err)
			}

			// Display QR code so users can easily add their keys to their
			// authenticator apps
			println("2FA has been enabled!")

			qr.New().Get(fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=Inertia",
				username, totpInfo.TotpSecret)).Print()
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
			if err := root.getUserClient().DisableTotp(root.context()); err != nil {
				output.Fatal(err)
			}
			println("2FA successfully disabled")
		},
	}
	root.AddCommand(disable)
}
