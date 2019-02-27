// +build !no_bootstrap

package bootstrap

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
			Port: "4303",
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
	assert.NoError(t, Bootstrap(c, Options{Out: os.Stdout}))

	// Daemon setup takes a bit of time - do a crude wait
	time.Sleep(3 * time.Second)

	// Check if daemon is online following bootstrap
	status, err := c.Status(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "test", status.InertiaVersion)
}
