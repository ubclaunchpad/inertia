package provisioncmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/client/bootstrap"
	"github.com/ubclaunchpad/inertia/client/runner"
	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/input"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
	"github.com/ubclaunchpad/inertia/provision"
)

// ProvisionCmd is the parent class for the 'inertia provision' subcommands
type ProvisionCmd struct {
	*cobra.Command
	remotes *cfg.Remotes
	project *cfg.Project
	cfgPath string
}

const (
	flagDaemonPort = "daemon.port"
	flagPorts      = "ports"
)

// AttachProvisionCmd attaches the 'provision' subcommands to the given parent
func AttachProvisionCmd(inertia *core.Cmd) {
	var prov = &ProvisionCmd{}
	prov.Command = &cobra.Command{
		Use:   "provision",
		Short: "Provision a new remote host to deploy your project on",
		Long:  `Provisions a new remote host set up for continuous deployment with Inertia.`,
		PersistentPreRun: func(*cobra.Command, []string) {
			var err error
			prov.remotes, err = local.GetRemotes()
			if err != nil {
				out.Fatalf(err.Error())
			}
			prov.project, err = local.GetProject(inertia.ProjectConfigPath)
			if err != nil {
				out.Fatalf(err.Error())
			}
			prov.cfgPath = inertia.ProjectConfigPath
		},
	}
	prov.PersistentFlags().StringP(flagDaemonPort, "d", "4303", "daemon port")
	prov.PersistentFlags().StringArrayP(flagPorts, "p", []string{}, "ports your project uses")

	// add children
	prov.attachEcsCmd()

	// add to parent
	inertia.AddCommand(prov.Command)
}

func (root *ProvisionCmd) attachEcsCmd() {
	const (
		flagType        = "type"
		flagUser        = "user"
		flagFromEnv     = "from-env"
		flagFromProfile = "from-profile"
		flagProfilePath = "profile.path"
		flagProfileUser = "profile.user"
	)
	var provEC2 = &cobra.Command{
		Use:   "ec2 [name]",
		Short: "[BETA] Provision a new Amazon EC2 instance",
		Long: `[BETA] Provisions a new Amazon EC2 instance and sets it up for continuous deployment
with Inertia. 

Make sure you run this command with the '-p' flag to indicate what ports
your project uses - for example:

	inertia provision ec2 my_ec2_instance -p 8000

This ensures that your project ports are properly exposed and externally accessible.
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if _, found := root.remotes.GetRemote(args[0]); found {
				out.Fatal("remote with name already exists")
			}

			// Load flags for setup configuration
			var (
				user, _               = cmd.Flags().GetString(flagUser)
				instanceType, _       = cmd.Flags().GetString(flagType)
				stringProjectPorts, _ = cmd.Flags().GetStringArray(flagPorts)
			)
			if stringProjectPorts == nil || len(stringProjectPorts) == 0 {
				out.Print(out.C("[WARNING] no project ports provided - this means that no ports"+
					"will be exposed on your ec2 host. Use the '--ports' flag to set"+
					"ports that you want to be accessible.\n", out.RD))
			}
			var highlight = out.NewColorer(out.CY)

			// Load flags for credentials
			var (
				fromEnv, _     = cmd.Flags().GetBool(flagFromEnv)
				withProfile, _ = cmd.Flags().GetBool(flagFromProfile)
			)

			// Create VPS instance
			var prov *provision.EC2Provisioner
			var err error
			if fromEnv {
				prov, err = provision.NewEC2ProvisionerFromEnv(user, os.Stdout)
				if err != nil {
					out.Fatal(err)
				}
			} else if withProfile {
				var profileUser, _ = cmd.Flags().GetString(flagProfileUser)
				var profilePath, _ = cmd.Flags().GetString(flagProfilePath)
				prov, err = provision.NewEC2ProvisionerFromProfile(
					user, profileUser, profilePath, os.Stdout)
				if err != nil {
					out.Fatal(err)
				}
			} else {
				keyID, key, err := enterEC2CredentialsWalkthrough()
				if err != nil {
					out.Fatal(err)
				}
				prov, err = provision.NewEC2Provisioner(user, keyID, key, os.Stdout)
				if err != nil {
					out.Fatal(err)
				}
			}

			// Report connected user
			out.Printf("Executing commands as user '%s'\n", prov.GetUser())

			// Prompt for region
			out.Println("See https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions for a list of available regions.")
			region, err := input.NewPrompt(nil).
				Prompt(highlight.S("Please enter a region: ")).
				GetString()
			if err != nil {
				out.Fatal(err)
			}

			// List image options and prompt for input
			out.Printf("Loading images for region '%s'...\n", region)
			images, err := prov.ListImageOptions(region)
			if err != nil {
				out.Fatal(err)
			}
			// allow arbitrary
			image, err := input.NewPrompt(&input.PromptConfig{AllowInvalid: true}).
				PromptFromList("image", images).
				GetString()
			if err != nil {
				out.Fatal(err)
			}

			// Gather input
			out.Printf("Creating %s instance in %s from image %s...\n", instanceType, region, image)
			var ports = []int64{}
			for _, portString := range stringProjectPorts {
				p, err := common.ParseInt64(portString)
				if err == nil {
					ports = append(ports, p)
				} else {
					out.Printf("invalid port %s", portString)
				}
			}

			// Create remote instance
			var port, _ = cmd.Flags().GetString(flagDaemonPort)
			var portDaemon, _ = common.ParseInt64(port)
			remote, err := prov.CreateInstance(provision.EC2CreateInstanceOptions{
				Name:        args[0],
				ProjectName: root.project.Name,
				Ports:       ports,
				DaemonPort:  portDaemon,

				ImageID:      image,
				InstanceType: instanceType,
				Region:       region,
			})
			if err != nil {
				out.Fatal(err)
			}
			out.Println(highlight.Sf("Instance provisioned for remote '%s'!", args[0]))

			// Save new remote to configuration
			local.SaveRemote(remote)

			// Create inertia client
			var inertia = client.NewClient(remote, client.Options{
				SSH: runner.SSHOptions{KeyPassphrase: os.Getenv(local.EnvSSHPassphrase)},
			})

			// Bootstrap remote
			out.Println(highlight.Sf("Initializing Inertia daemon at %s...", inertia.Remote.IP))
			var repo = common.ExtractRepository(common.GetSSHRemoteURL(root.project.URL))
			if err := bootstrap.Bootstrap(inertia, bootstrap.Options{
				RepoName: repo,
				Out:      os.Stdout,
			}); err != nil {
				out.Fatal(err.Error())
			}

			// Save updated config
			out.Println(highlight.S("Saving remote..."))
			if err := local.SaveRemote(remote); err != nil {
				out.Fatal(err.Error())
			}
		},
	}
	provEC2.Flags().StringP(flagType, "t",
		"t2.micro", "ec2 instance type to instantiate")
	provEC2.Flags().StringP(flagUser, "u",
		"ec2-user", "ec2 instance user to execute commands as")
	provEC2.Flags().Bool(flagFromEnv, false,
		"load ec2 credentials from environment - requires AWS_ACCESS_KEY_ID, AWS_ACCESS_KEY to be set")
	provEC2.Flags().Bool(flagFromProfile, false,
		"load ec2 credentials from profile")
	provEC2.Flags().String(flagProfilePath, "~/.aws/credentials",
		"path to aws profile credentials file")
	provEC2.Flags().String(flagProfileUser, "default",
		"user profile for aws credentials file")

	root.AddCommand(provEC2)
}
