package bootstrap

import (
	"fmt"
	"io"

	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/common"
)

// Options denotes configuration for the bootstrapping process.
// RepoName is optional, and only used for generating printed links.
// Out is where output will be written.
type Options struct {
	RepoName string
	Out      io.Writer
}

// Bootstrap bootstraps the given remote
func Bootstrap(c *client.Client, opts Options) error {
	var out = opts.Out
	if out == nil {
		out = &common.DevNull{}
	}
	sshc, err := c.GetSSHClient()
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Setting up remote at %s\n", c.Remote.IP)
	fmt.Fprint(out, ">> Step 1/4: Installing docker...\n")
	if err := sshc.InstallDocker(); err != nil {
		return err
	}

	fmt.Fprint(out, ">> Step 2/4: Building deploy key...\n")
	pub, err := sshc.GenerateKeys()
	if err != nil {
		return err
	}

	// This step needs to run before any other commands that rely on
	// the daemon image, since the daemon is loaded here.
	fmt.Fprint(out, ">> Step 3/4: Starting daemon...\n")
	if err = sshc.DaemonUp(); err != nil {
		return err
	}

	fmt.Fprint(out, ">> Step 4/4: Fetching daemon API token...\n")
	if err := sshc.AssignAPIToken(); err != nil {
		return err
	}

	fmt.Fprintf(out, `Inertia has been set up and daemon is running on remote!

You may have to wait briefly for Inertia to set up some dependencies.
Use 'inertia %s logs' to check on the daemon's setup progress.
`, c.Remote.Name)

	// pretty divider
	fmt.Fprint(out, "=============================\n\n")

	// Output deploy key to user
	fmt.Fprintf(out, ">> GitHub Deploy Key (add to https://www.github.com/%s/settings/keys/new):\n",
		c.Remote.Name)
	fmt.Fprint(out, pub+"\n")

	// Output Webhook url to user
	var addr, _ = c.Remote.DaemonAddr()
	fmt.Fprintf(out, `
>> GitHub WebHook URL (add to https://www.github.com/%s/settings/hooks/new):
Address:  https://%s/webhook
Secret:   %s
Note that by default, you will have to disable SSL verification in your webhook
settings - Inertia uses self-signed certificates that GitHub won't be able to
verify.`, opts.RepoName, addr, c.Remote.Daemon.WebHookSecret)

	fmt.Fprintf(out, `
Inertia daemon successfully deployed! Add your webhook url and deploy key to
your repository to enable continuous deployment.

Then run 'inertia %s up' to deploy your application.
`, c.Remote.Name)
	return nil
}
