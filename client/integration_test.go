package client

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBootstrap_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cli := getIntegrationClient(nil)
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
