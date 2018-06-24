package main

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
}

var cmdDeploymentEnvSet = &cobra.Command{
	Use:   "set [NAME] [VALUE]",
	Short: "Set an environment variable on your remote",
	Long: `Set a persistent environment variable on your remote. Set environment
variables are applied to all deployed containers.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, err := local.GetClient(remoteName, ConfigFilePath, cmd)
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
	Use:   "rm [NAME]",
	Short: "Remove an environment variable from your remote",
	Long: `Remove the specified environment variable from deployed containers
and persistent environment storage.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, err := local.GetClient(remoteName, ConfigFilePath, cmd)
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
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := strings.Split(cmd.Parent().Parent().Use, " ")[0]
		deployment, err := local.GetClient(remoteName, ConfigFilePath, cmd)
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
