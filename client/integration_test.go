package client

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/cfg"
)

func newIntegrationClient(mockRunner *mockSSHRunner) *Client {
	remote := &cfg.RemoteVPS{
		IP:      "127.0.0.1",
		PEM:     "../test/keys/id_rsa",
		User:    "root",
		SSHPort: "69",
		Daemon: &cfg.DaemonConfig{
			Port: "4303",
		},
	}
	if mockRunner != nil {
		mockRunner.r = remote
		return &Client{
			version:   "test",
			RemoteVPS: remote,
			out:       os.Stdout,
			sshRunner: mockRunner,
		}
	}
	return &Client{
		version:   "test",
		RemoteVPS: remote,
		out:       os.Stdout,
		sshRunner: NewSSHRunner(remote),
	}
}

func TestBootstrap_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cli := newIntegrationClient(nil)
	err := cli.BootstrapRemote("")
	assert.Nil(t, err)

	// Daemon setup takes a bit of time - do a crude wait
	time.Sleep(3 * time.Second)

	// Check if daemon is online following bootstrap
	host := "https://" + cli.GetIPAndPort()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(host)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
}
