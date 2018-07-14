package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/local"
)

var cmdDeploymentEnv = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables on your remote",
	Long: `Manages environment variables on your remote through Inertia. 
	
Configured variables can be encrypted or stored in plain text, and are applied to 
all project containers on startup.`,
}

var cmdDeploymentEnvSet = &cobra.Command{
	Use:   "set [name] [value]",
	Short: "Set an environment variable on your remote",
	Long: `Sets a persistent environment variable on your remote. Set environment
variables are applied to all deployed containers.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		encrypt, err := cmd.Flags().GetBool("encrypt")
		if err != nil {
			log.Fatal(err)
		}

		resp, err := deployment.UpdateEnv(args[0], args[1], encrypt, false)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
	},
}

var cmdDeploymentEnvRemove = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove an environment variable from your remote",
	Long: `Removes the specified environment variable from deployed containers
and persistent environment storage.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := deployment.UpdateEnv(args[0], "", false, true)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
	},
}

var cmdDeploymentEnvList = &cobra.Command{
	Use:   "ls",
	Short: "List currently set and saved environment variables",
	Long: `Lists currently set and saved environment variables. The values of encrypted
variables are not be decrypted.`,
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, _, err := local.GetClient(remoteName, configFilePath, cmd)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := deployment.ListEnv()
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
	},
}
