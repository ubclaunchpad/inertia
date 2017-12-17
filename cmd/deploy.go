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
	"log"
	"os"

	"github.com/spf13/cobra"
)

// TODO: Reference daemon pkg for this information?
// We only want the package dependencies to go in one
// direction, so best to think about how to do this.
// Clearly cannot ask for this information over HTTP.
var defaultDaemonPort = "8081"

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy [REMOTE]",
	Short: "Start continuous deployment to the remote VPS instance specified",
	Long: `Start continuous deployment to the remote VPS instance specified.
Run 'inertia remote status' beforehand to ensure your daemon is running.
Requires:

1. A deploy key to be registered for the daemon with your GitHub repository.
2. A webhook url to registered for the daemon with your GitHub repository.

Run 'inertia remote bootstrap [REMOTE]' to collect these.`,
	Args: cobra.MinimumNArgs(1),
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

		stop, err := cmd.Flags().GetBool("stop")
		if err != nil {
			log.Fatal(err)
		}

		if stop {
			// Stop the deployment.
			println("stop not currently implemented")

		} else {
			// Start the deployment
			err = config.CurrentRemoteVPS.Deploy()
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	deployCmd.Flags().BoolP("stop", "s", false, "Stop the deployment")
}

// Deploy deploys the project to the remote VPS instance specified
// in the configuration object.
func (remote *RemoteVPS) Deploy() error {
	return nil
}
