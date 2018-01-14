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

// initCmd represents the init command
import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/client"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an inertia project in this repository",
	Long: `Initialize an inertia project in this GitHub repository.
There must be a local git repository in order for initialization
to succeed.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := client.InitializeInertiaProject()
		if err != nil {
			log.Fatal(err)
		}
		println("A .inertia folder has been created to store Inertia configuration.")
		println("It is recommended that you DO NOT commit this folder in source")
		println("control since it will be used to store sensitive information.")
		println("\nYou can now use 'inertia remote add' to connect your remote")
		println("VPS instance.")
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
