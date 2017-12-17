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
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// TODO: Reference daemon pkg for this information?
// We only want the package dependencies to go in one
// direction, so best to think about how to do this.
// Clearly cannot ask for this information over HTTP.
var defaultDaemonPort = "8081"

const (
	daemonUp   = "up"
	daemonDown = "down"
)

// DaemonRequester can make HTTP requests to the daemon.
type DaemonRequester interface {
	Up() error
	Down() error
}

// Deployment manages a deployment and implements the
// DaemonRequester interface.
type Deployment struct {
	*RemoteVPS
	Repository string
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
			log.Fatal(err)
		}

		if args[0] != config.CurrentRemoteName {
			println("No such remote " + args[0])
			println("Inertia currently supports one remote per repository")
			println("Run `inertia remote -v' to see what remote is available")
			os.Exit(1)
		}

		switch args[1] {
		case daemonUp:
			// Start the deployment
			deployment := &Deployment{
				RemoteVPS:  config.CurrentRemoteVPS,
				Repository: "",
			}
			err := deployment.Up()
			if err != nil {
				log.Fatal(err)
			}
		case daemonDown:
			// Start the deployment
			deployment := &Deployment{
				RemoteVPS:  config.CurrentRemoteVPS,
				Repository: "",
			}
			err := deployment.Down()
			if err != nil {
				log.Fatal(err)
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
// in the configuration object.
func (d *Deployment) Up() error {
	return nil
}

// Down brings the project down on the remote VPS instance specified
// in the configuration object.
func (d *Deployment) Down() error {
	return nil
}
