package hostcmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cmd/printutil"
)

type EnvCmd struct {
	*cobra.Command
	host *HostCmd
}

func AttachEnvCmd(host *HostCmd) {
	var env = &EnvCmd{
		Command: &cobra.Command{
			Use:   "env",
			Short: "Manage environment variables on your remote",
			Long: `Manages environment variables on your remote through Inertia. 
			
Configured variables can be encrypted or stored in plain text. They are applied
as follows:

- for docker-compose projects, variables are set for the docker-compose process
- for Dockerfile projects, variables are set in the deployed container
`,
		},
		host: host,
	}

	// attach children
	env.attachSetCmd()
	env.attachListCmd()
	env.attachRemoveCmd()

	// attach to parent
	host.AddCommand(env.Command)
}

func (root *EnvCmd) attachSetCmd() {
	var set = &cobra.Command{
		Use:   "set [name] [value]",
		Short: "Set an environment variable on your remote",
		Long: `Sets a persistent environment variable on your remote. Set environment
variables are applied to all deployed containers.`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var encrypt, _ = cmd.Flags().GetBool("encrypt")
			resp, err := root.host.client.UpdateEnv(args[0], args[1], encrypt, false)
			if err != nil {
				printutil.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				printutil.Fatal(err)
			}
			fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
		},
	}
	set.Flags().BoolP("encrypt", "e", false, "encrypt variable when stored")
	root.AddCommand(set)
}

func (root *EnvCmd) attachRemoveCmd() {
	var remove = &cobra.Command{
		Use:   "rm [name]",
		Short: "Remove an environment variable from your remote",
		Long: `Removes the specified environment variable from deployed containers
and persistent environment storage.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := root.host.client.UpdateEnv(args[0], "", false, true)
			if err != nil {
				printutil.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				printutil.Fatal(err)
			}

			fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
		},
	}
	root.AddCommand(remove)
}

func (root *EnvCmd) attachListCmd() {
	var list = &cobra.Command{
		Use:   "ls",
		Short: "List currently set and saved environment variables",
		Long: `Lists currently set and saved environment variables. The values of encrypted
variables are not be decrypted.`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := root.host.client.ListEnv()
			if err != nil {
				printutil.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				printutil.Fatal(err)
			}
			fmt.Printf("(Status code %d) %s\n", resp.StatusCode, body)
		},
	}
	root.AddCommand(list)
}
