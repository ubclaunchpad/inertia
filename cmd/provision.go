package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
	"github.com/ubclaunchpad/inertia/provision"
)

// Initialize "inertia" commands regarding basic configuration
func init() {
	cmdProvisionECS.Flags().StringP(
		"type", "t", "t2.micro", "ec2 instance type to instantiate",
	)
	cmdProvisionECS.Flags().StringP(
		"user", "u", "ec2-user", "ec2 instance user to execute commands as",
	)
	cmdProvisionECS.Flags().String(
		"profile-path", "", "path to aws credentials file",
	)
	cmdProvisionECS.Flags().String(
		"profile-user", "", "user profile for aws credentials file",
	)
	cmdProvisionECS.Flags().Bool(
		"from-env", false, "load ec2 credentials from environment - requires AWS_ACCESS_KEY_ID, AWS_ACCESS_KEY to be set",
	)
	cmdProvision.AddCommand(cmdProvisionECS)
	cmdProvision.PersistentFlags().StringP("daemon-port", "d", "4303", "daemon port")
	cmdProvision.PersistentFlags().StringArrayP("ports", "p", []string{}, "ports your project uses")
	Root.AddCommand(cmdProvision)
}

var cmdProvision = &cobra.Command{
	Use:   "provision",
	Short: "Provision a new remote host to deploy your project on",
	Long:  `Provisions a new remote host set up for continuous deployment with Inertia.`,
}

var cmdProvisionECS = &cobra.Command{
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
		// Ensure project initialized.
		config, path, err := local.GetProjectConfigFromDisk(configFilePath)
		if err != nil {
			log.Fatal(err)
		}
		if _, found := config.GetRemote(args[0]); found {
			log.Fatal("remote with name already exists")
		}

		// Load flags for credentials
		var fromEnv, _ = cmd.Flags().GetBool("from-env")
		var profilePath, _ = cmd.Flags().GetString("profile-path")

		// Load flags for setup configuration
		var user, _ = cmd.Flags().GetString("user")
		var instanceType, _ = cmd.Flags().GetString("type")
		var stringProjectPorts, _ = cmd.Flags().GetStringArray("ports")

		if stringProjectPorts == nil || len(stringProjectPorts) == 0 {
			fmt.Print("[WARNING] no project ports provided - this means that no ports" +
				"will be exposed on your ec2 host. Use the '--ports' flag to set" +
				"ports that you want to be accessible.")
		}

		// Create VPS instance
		var prov *provision.EC2Provisioner
		if fromEnv {
			prov, err = provision.NewEC2ProvisionerFromEnv(user, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		} else if profilePath != "" {
			var profileUser, _ = cmd.Flags().GetString("profile-user")
			prov, err = provision.NewEC2ProvisionerFromProfile(
				user, profileUser, profilePath, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			keyID, key, err := enterEC2CredentialsWalkthrough(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
			prov, err = provision.NewEC2Provisioner(user, keyID, key, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Report connected user
		fmt.Printf("Executing commands as user '%s'\n", prov.GetUser())

		// Prompt for region
		println("See https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions for a list of available regions.")
		print("Please enter a region: ")
		var region string
		if _, err = fmt.Fscanln(os.Stdin, &region); err != nil {
			log.Fatal(err)
		}

		// List image options and prompt for input
		fmt.Printf("Loading images for region '%s'...\n", region)
		images, err := prov.ListImageOptions(region)
		if err != nil {
			log.Fatal(err)
		}
		image, err := chooseFromListWalkthrough(os.Stdin, "image", images)
		if err != nil {
			log.Fatal(err)
		}

		// Gather input
		fmt.Printf("Creating %s instance in %s from image %s...\n", instanceType, region, image)
		var ports = []int64{}
		for _, portString := range stringProjectPorts {
			p, err := common.ParseInt64(portString)
			if err == nil {
				ports = append(ports, p)
			} else {
				fmt.Printf("invalid port %s", portString)
			}
		}

		// Create remote instance
		var port, _ = cmd.Flags().GetString("daemon-port")
		var portDaemon, _ = common.ParseInt64(port)
		remote, err := prov.CreateInstance(provision.EC2CreateInstanceOptions{
			Name:        args[0],
			ProjectName: config.Project,
			Ports:       ports,
			DaemonPort:  portDaemon,

			ImageID:      image,
			InstanceType: instanceType,
			Region:       region,
		})
		if err != nil {
			log.Fatal(err)
		}

		// Save new remote to configuration
		remote.Branch, err = local.GetRepoCurrentBranch()
		if err != nil {
			log.Fatal(err)
		}
		config.AddRemote(remote)
		config.Write(path)

		// Create inertia client
		inertia, found := client.NewClient(args[0], config, os.Stdout)
		if !found {
			log.Fatal("vps setup did not complete properly")
		}

		// Bootstrap remote
		fmt.Printf("Initializing Inertia daemon at %s...\n", inertia.RemoteVPS.IP)
		if err = inertia.BootstrapRemote(config.Project); err != nil {
			log.Fatal(err)
		}

		// Save updated config
		config.Write(path)
	},
}
