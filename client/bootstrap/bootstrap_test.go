// +build !no_bootstrap

package bootstrap

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/client/runner"
)

func newIntegrationClient() *client.Client {
	remote := &cfg.Remote{
		Version: "test",
		IP:      "127.0.0.1",
		SSH: &cfg.SSH{
			IdentityFile: "../../test/keys/id_rsa",
			User:         "root",
			SSHPort:      "69",
		},
		Daemon: &cfg.Daemon{
			Port:          "4303",
			WebHookSecret: "sekret",
		},
	}
	return client.NewClient(remote, client.Options{
		SSH:   runner.SSHOptions{},
		Out:   os.Stdout,
		Debug: true,
	})
}

func TestBootstrap_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var c = newIntegrationClient()
	c.WithDebug(true) // makes troubleshooting tests easier
	require.NoError(t, Bootstrap(c, Options{Out: os.Stdout}), "bootstrap failed")

	// Daemon setup takes a bit of time - do a crude wait
	time.Sleep(5 * time.Second)

	// Check if daemon is online following bootstrap
	status, err := c.Status(context.Background())
	require.NoError(t, err, "status check of bootstrapped daemon failed")
	assert.Equal(t, "test", status.InertiaVersion)
}
