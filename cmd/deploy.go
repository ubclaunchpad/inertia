// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the application to the remote VPS instance specified",
	Long: `Deploy the application to the remote VPS instance specified.
A URL will be provided to direct GitHub webhooks too, the daemon will
request access to the repository via a public key, the daemon will begin
waiting for updates to this repository's remote master branch.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := GetProjectConfigFromDisk()
		config.CurrentRemoteVPS.Deploy()
	},
}

// Deploy deploys the project to the remote VPS instance specified
// in the configuration object.
func (remote *RemoteVPS) Deploy() {
	println("Deploying remote")
	println("Installing Docker on remote instance...")
	result, _ := remote.RunSSHScript("bootstrap/docker.sh")
	print(string(result.Bytes()))

	println("Running Inertia daemon on remote instance...")
	// remote.RunSSHCommand("docker run ubclaunchpad/inertia")
}

func init() {
	RootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().StringP("user", "u", "root", "Set the user for SSH access")
}
