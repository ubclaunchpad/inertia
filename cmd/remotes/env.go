package remotescmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/output"
)

// EnvCmd is the parent class for the 'env' subcommands
type EnvCmd struct {
	*cobra.Command
	host *HostCmd
}

// AttachEnvCmd attaches the 'env' subcommands to the given host
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

// Context returns the root host command's context
func (root *EnvCmd) Context() context.Context { return root.host.ctx }

func (root *EnvCmd) attachSetCmd() {
	const flagEncrypt = "encrypt"
	var set = &cobra.Command{
		Use:   "set [name] [value]",
		Short: "Set an environment variable on your remote",
		Long: `Sets a persistent environment variable on your remote. Set environment
variables are applied to all deployed containers.`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var encrypt, _ = cmd.Flags().GetBool(flagEncrypt)
			if err := root.host.client.UpdateEnv(
				root.Context(),
				args[0],
				args[1],
				encrypt,
				false,
			); err != nil {
				output.Fatal(err)
			}
			println("env value successfully updated")
		},
	}
	set.Flags().BoolP(flagEncrypt, "e", false, "encrypt variable when stored")
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
			if err := root.host.client.UpdateEnv(
				root.Context(),
				args[0],
				"",
				false,
				true,
			); err != nil {
				output.Fatal(err)
			}
			println("env value successfully removed")
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
			variables, err := root.host.client.ListEnv(root.Context())
			if err != nil {
				output.Fatal(err)
			}

			if len(variables) == 0 {
				println("no variables configured on remote")
			} else {
				println(strings.Join(variables, "\n"))
			}
		},
	}
	root.AddCommand(list)
}
