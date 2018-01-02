// Copyright © 2017 UBC Launch Pad team@ubclaunchpad.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
)

// TODO: Reference daemon pkg for this information?
// We only want the package dependencies to go in one
// direction, so best to think about how to do this.
// Clearly cannot ask for this information over HTTP.
var defaultDaemonPort = "8081"

const (
	daemonUp     = "up"
	daemonDown   = "down"
	daemonStatus = "status"
)

// DaemonRequester can make HTTP requests to the daemon.
type DaemonRequester interface {
	Up() (*http.Response, error)
	Down() (*http.Response, error)
}

// UpRequest is the body of a up request to the daemon.
type UpRequest struct {
	Repo string `json:"repo"`
}

// Deployment manages a deployment and implements the
// DaemonRequester interface.
type Deployment struct {
	*RemoteVPS
	Repository *git.Repository
}

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy [REMOTE] [COMMAND]",
	Short: "Configure continuous deployment to the remote VPS instance specified",
	Long: `Start or stop continuous deployment to the remote VPS instance specified.
Run 'inertia remote status' beforehand to ensure your daemon is running.
Requires:

1. A deploy key to be registered for the daemon with your GitHub repository.
2. A webhook url to registered for the daemon with your GitHub repository.

Run 'inertia remote bootstrap [REMOTE]' to collect these.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		config, err := GetProjectConfigFromDisk()
		if err != nil {
			log.WithError(err)
		}

		if args[0] != config.CurrentRemoteName {
			println("No such remote " + args[0])
			println("Inertia currently supports one remote per repository")
			println("Run `inertia remote -v' to see what remote is available")
			os.Exit(1)
		}

		repo, err := getLocalRepo()
		if err != nil {
			log.WithError(err)
		}

		deployment := &Deployment{
			RemoteVPS:  config.CurrentRemoteVPS,
			Repository: repo,
		}

		switch args[1] {
		case daemonUp:
			// Start the deployment
			resp, err := deployment.Up()
			if err != nil {
				log.Fatal(err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.WithError(err)
			}

			switch resp.StatusCode {
			case http.StatusCreated:
				fmt.Printf("(Status code %d) Project up\n", resp.StatusCode)
			case http.StatusForbidden:
				fmt.Printf("(Status code %d) Bad auth: %s\n", resp.StatusCode, body)
			default:
				fmt.Printf("(Status code %d) Unknown response from daemon: %s",
					resp.StatusCode, body)
			}

		case daemonDown:
			// Shut down the deployment
			resp, err := deployment.Down()
			if err != nil {
				log.WithError(err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.WithError(err)
			}

			switch resp.StatusCode {
			case http.StatusOK:
				fmt.Printf("(Status code %d) Project down\n", resp.StatusCode)
			case http.StatusPreconditionFailed:
				fmt.Printf("(Status code %d) No containers are currently active\n", resp.StatusCode)
			case http.StatusForbidden:
				fmt.Printf("(Status code %d) Bad auth: %s\n", resp.StatusCode, body)
			default:
				fmt.Printf("(Status code %d) Unknown response from daemon: %s\n",
					resp.StatusCode, body)
			}

		case daemonStatus:
			// Get status of the deployment
			resp, err := deployment.Status()
			if err != nil {
				log.WithError(err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.WithError(err)
			}

			switch resp.StatusCode {
			case http.StatusOK:
				fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
			case http.StatusForbidden:
				fmt.Printf("(Status code %d) Bad auth: %s\n", resp.StatusCode, body)
			default:
				fmt.Printf("Unknown response from daemon: %d %s\n",
					resp.StatusCode, body)
			}

		default:
			fmt.Printf("No such deployment command: %s\n", args[1])
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
}

// Up brings the project up on the remote VPS instance specified
// in the deployment object.
func (d *Deployment) Up() (*http.Response, error) {
	host := "http://" + d.RemoteVPS.GetIPAndPort() + "/up"
	repo, err := getLocalRepo()
	if err != nil {
		return nil, err
	}

	// TODO: Support other repo names.
	origin, err := repo.Remote("origin")
	if err != nil {
		return nil, err
	}

	req := UpRequest{Repo: getSSHRemoteURL(origin.Config().URLs[0])}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(host, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.New("Error when deploying project")
	}

	return resp, nil
}

// Down brings the project down on the remote VPS instance specified
// in the configuration object.
func (d *Deployment) Down() (*http.Response, error) {
	host := "http://" + d.RemoteVPS.GetIPAndPort() + "/down"
	resp, err := http.Post(host, "application/json", nil)
	if err != nil {
		return nil, errors.New("Error when shutting down project")
	}
	return resp, nil
}

// Status lists the currently active containers on the remote VPS instance
func (d *Deployment) Status() (*http.Response, error) {
	host := "http://" + d.RemoteVPS.GetIPAndPort() + "/status"
	resp, err := http.Post(host, "application/json", nil)
	if err != nil {
		return nil, errors.New("Error when checking project status")
	}
	return resp, nil
}
