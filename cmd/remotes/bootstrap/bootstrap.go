package bootstrap

import (
	"fmt"

	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/common"
)

func SetUpRemote(name, gitURL string, c *client.Client) error {
	sshc, err := c.GetSSHClient()
	if err != nil {
		return err
	}

	fmt.Printf("Setting up remote at %s\n", c.Remote.IP)
	fmt.Print(">> Step 1/4: Installing docker...\n")
	if err := sshc.InstallDocker(); err != nil {
		return err
	}

	fmt.Print(">> Step 2/4: Building deploy key...\n")
	pub, err := sshc.GenerateKeys()
	if err != nil {
		return err
	}

	// This step needs to run before any other commands that rely on
	// the daemon image, since the daemon is loaded here.
	fmt.Print(">> Step 3/4: Starting daemon...\n")
	if err = sshc.DaemonUp(); err != nil {
		return err
	}

	fmt.Print(">> Step 4/4: Fetching daemon API token...\n")
	if err := sshc.AssignAPIToken(); err != nil {
		return err
	}

	fmt.Printf(`Inertia has been set up and daemon is running on remote!

You may have to wait briefly for Inertia to set up some dependencies.
Use 'inertia %s logs' to check on the daemon's setup progress.`, name)

	// pretty divider
	fmt.Print("=============================\n\n")

	// get repo name for pretty printing
	var repo = common.ExtractRepository(common.GetSSHRemoteURL(gitURL))

	// Output deploy key to user
	fmt.Printf(">> GitHub Deploy Key (add to https://www.github.com/%s/settings/keys/new):\n", repo)
	fmt.Print(pub.String() + "\n")

	// Output Webhook url to user
	var addr, _ = c.Remote.GetDaemonAddr()
	fmt.Printf(`
>> GitHub WebHook URL (add to https://www.github.com/%s/settings/hooks/new):
Address:  https://%s/webhook
Secret:   %s
Note that by default, you will have to disable SSL verification in your webhook
settings - Inertia uses self-signed certificates that GitHub won't be able to
verify.`, repo, addr, c.Remote.Daemon.WebHookSecret)

	fmt.Printf(`
Inertia daemon successfully deployed! Add your webhook url and deploy key to
your repository to enable continuous deployment.

Then run 'inertia %s up' to deploy your application.`, name)
	return nil
}
