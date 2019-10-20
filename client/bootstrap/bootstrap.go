package bootstrap

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji"

	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/common"
)

// Options denotes configuration for the bootstrapping process.
// RepoName is optional, and only used for generating printed links.
// Out is where output will be written.
type Options struct {
	RepoName     string
	DisableColor bool
	DisableEmoji bool
	Out          io.Writer
}

// Bootstrap bootstraps the given remote
func Bootstrap(c *client.Client, opts Options) error {
	// The bootstrap script is a bit of an outlier, as it is separated from cmd/...
	// but it does a lot of printing. we use emoji and color directly in this
	// script to avoid introducing a direct dependency on cmd/...
	color.Output = opts.Out
	var highlight = color.New(color.FgYellow, color.Bold)
	var blue = color.New(color.FgBlue, color.Bold)
	if opts.DisableColor {
		highlight.DisableColor()
		blue.DisableColor()
	}
	var fprintf = emoji.Fprintf
	if opts.DisableEmoji {
		fprintf = fmt.Fprintf
	}

	var out = opts.Out
	if out == nil {
		out = &common.DevNull{}
	}
	sshc, err := c.GetSSHClient()
	if err != nil {
		return fmt.Errorf("failed to initialize SSH client: %w", err)
	}

	fprintf(out, "Setting up remote '%s' at %s\n", c.Remote.Name, c.Remote.IP)
	emoji.Fprint(out, ":whale: ")
	blue.Fprint(out, "Step 1/4: Installing docker...\n")
	if err := sshc.InstallDocker(); err != nil {
		return err
	}

	emoji.Fprint(out, ":hammer_and_wrench: ")
	blue.Fprint(out, "Step 2/4: Building deploy key...\n")
	pub, err := sshc.GenerateKeys()
	if err != nil {
		return err
	}

	// This step needs to run before any other commands that rely on
	// the daemon image, since the daemon is loaded here.
	emoji.Fprint(out, ":robot: ")
	blue.Fprint(out, "Step 3/4: Starting daemon...\n")
	if err = sshc.DaemonUp(); err != nil {
		return err
	}

	emoji.Fprint(out, ":lock: ")
	blue.Fprint(out, "Step 4/4: Fetching daemon API token...\n")
	if err := sshc.AssignAPIToken(); err != nil {
		return err
	}

	fmt.Fprintf(out, `Inertia has been set up and daemon is running on remote!

You may have to wait briefly for Inertia to set up some dependencies.
Use 'inertia %s logs' to check on the daemon's setup progress.
`, c.Remote.Name)

	// pretty divider
	fmt.Fprint(out, "\n==========================================================\n\n")

	// Output deploy key to user
	fprintf(out, ":star: ")
	highlight.Fprintf(out,
		"GitHub Deploy Key (add to https://www.github.com/%s/settings/keys/new):\n",
		opts.RepoName)
	fmt.Fprint(out, pub+"\n")

	// Output Webhook url to user
	var addr, _ = c.Remote.DaemonAddr()
	fprintf(out, ":star: ")
	highlight.Fprintf(out,
		"GitHub WebHook URL (add to https://www.github.com/%s/settings/hooks/new)\n",
		opts.RepoName)
	fprintf(out, `:globe_with_meridians: Address:  %s/webhook
:key: Secret:   %s
Note that by default, you will have to disable SSL verification in your webhook
settings - Inertia uses self-signed certificates that GitHub won't be able to
verify. Read more about it here: https://inertia.ubclaunchpad.com/#custom-ssl-certificate
`, addr, c.Remote.Daemon.WebHookSecret)

	// pretty divider
	fmt.Fprint(out, "\n==========================================================\n")

	fprintf(out, `
Your Inertia daemon has been successfully deployed! Add your webhook url and
deploy key to your repository to enable continuous deployment. :rocket: 

Then run 'inertia %s up' to deploy your application!
`, c.Remote.Name)
	return nil
}
