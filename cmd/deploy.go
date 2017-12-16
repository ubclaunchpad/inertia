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

	"github.com/spf13/cobra"
)

// TODO: Reference daemon pkg for this information?
// We only want the package dependencies to go in one
// direction, so best to think about how to do this.
// Clearly cannot ask for this information over HTTP.
var defaultDaemonPort = "8081"

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the application to the remote VPS instance specified",
	Long: `Deploy the application to the remote VPS instance specified.
A URL will be provided to direct GitHub webhooks too, the daemon will
request access to the repository via a public key, the daemon will begin
waiting for updates to this repository's remote master branch.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		daemonPort, err := cmd.Flags().GetString("port")
		config.CurrentRemoteVPS.Deploy(config.CurrentRemoteName, daemonPort)
	},
}

// Deploy deploys the project to the remote VPS instance specified
// in the configuration object.
func (remote *RemoteVPS) Deploy(name, daemonPort string) {
	println("Deploying remote " + name)

	// Collect assets (docker shell script)
	installDockerSh, err := Asset("cmd/bootstrap/docker.sh")
	if err != nil {
		log.Fatal("Bootstrapping asset failed to load")
	}

	// Collect assets (keygen shell script)
	keygenSh, err := Asset("cmd/bootstrap/keygen.sh")
	if err != nil {
		log.Fatal("Bootstrapping asset failed to load")
	}

	// Install docker.
	_, stderr, err := remote.RunSSHCommand(string(installDockerSh))
	if err != nil {
		log.Println(stderr)
		log.Fatal("Failed to install docker on remote")
	}

	// Run inertia daemon (TODO).
	// remote.RunSSHCommand("docker run ubclaunchpad/inertia deamon run -p " + daemonPort)
	println("Daemon running on instance")

	// Create deploy key.
	result, stderr, err := remote.RunSSHCommand(string(keygenSh))

	if err != nil {
		log.Println(stderr)
		log.Fatal("Failed to run keygen on remote")
	}

	println()

	// Output deploy key to user.
	println("GitHub Deploy Key (add here https://www.github.com/<your_repo>/settings/hooks/new): ")
	println(string(result.Bytes()))

	// Output Webhook url to user.
	println("GitHub WebHook URL (add here https://www.github.com/<your_repo>/settings/keys/new): ")
	println("https://" + remote.IP + string(defaultDaemonPort))

	println()

	println("Inertia daemon successfully deployed, add webhook url and deploy key to enable it.")
}

func init() {
	RootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	deployCmd.Flags().StringP("port", "p", defaultDaemonPort, "Set the daemon port")
}
