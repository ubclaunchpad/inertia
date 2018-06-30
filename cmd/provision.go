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
		"type", "t", "t2.micro", "The ec2 instance type to instantiate",
	)
	cmdProvisionECS.Flags().StringP(
		"user", "u", "ec2-user", "The ec2 instance type to instantiate",
	)
	cmdProvisionECS.Flags().Bool(
		"from-env", false, "Load ec2 credentials from environment - requires AWS_ACCESS_KEY_ID, AWS_ACCESS_KEY to be set.",
	)
	cmdProvision.AddCommand(cmdProvisionECS)
	cmdProvision.PersistentFlags().StringP("daemon-port", "d", "4303", "Daemon port")
	cmdProvision.PersistentFlags().StringArrayP("ports", "p", []string{}, "Ports your project uses")
	Root.AddCommand(cmdProvision)
}

var cmdProvision = &cobra.Command{
	Use:   "provision",
	Short: "[BETA] Provision a new VPS setup for Inertia",
	Long:  `[BETA] Provision a new VPS instance set up for continuous deployment with Inertia.`,
}

var cmdProvisionECS = &cobra.Command{
	Use:   "ec2 [name]",
	Short: "[BETA] Provision a new Amazon EC2 instance",
	Long: `[BETA] Provision a new Amazon EC2 instance and set it up for continuous deployment
	with Inertia. Make sure you run this command with the '-p' flag to indicate what ports
	your project uses, since they must be exposed on your new instance.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, path, err := local.GetProjectConfigFromDisk(configFilePath)
		if err != nil {
			log.Fatal(err)
		}
		_, found := config.GetRemote(args[0])
		if found {
			log.Fatal("remote with name already exists")
		}

		// Load flags
		fromEnv, _ := cmd.Flags().GetBool("from-env")
		instanceType, _ := cmd.Flags().GetString("type")
		user, _ := cmd.Flags().GetString("user")
		stringProjectPorts, _ := cmd.Flags().GetStringArray("ports")

		// Create VPS instance
		var prov *provision.EC2Provisioner
		if !fromEnv {
			id, key, err := enterEC2CredentialsWalkthrough(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
			prov, err = provision.NewEC2Provisioner(id, key)
		} else {
			prov, err = provision.NewEC2ProvisionerFromEnv()
		}
		if err != nil {
			log.Fatal(err)
		}

		// Report connected user
		println("Successfully authenticated with user " + prov.GetUser())

		// Prompt for region
		println("See https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions for a list of available regions.")
		print("Please enter a region: ")
		var region string
		_, err = fmt.Fscanln(os.Stdin, &region)
		if err != nil {
			log.Fatal(err)
		}

		// List image options and prompt for input
		println("Loading images...")
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
		ports := []int64{}
		for _, portString := range stringProjectPorts {
			p, err := common.ParseInt64(portString)
			if err == nil {
				ports = append(ports, p)
			} else {
				fmt.Printf("invalid port %s", portString)
			}
		}
		port, _ := cmd.Flags().GetString("daemon-port")
		portDaemon, _ := common.ParseInt64(port)

		// Create remote instance
		remote, err := prov.CreateInstance(provision.EC2CreateInstanceOptions{
			Name:        args[0],
			ProjectName: config.Project,
			Ports:       ports,
			DaemonPort:  portDaemon,

			ImageID:      image,
			InstanceType: instanceType,
			Region:       region,

			User: user,
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
		inertia, found := client.NewClient(args[0], config)
		if !found {
			log.Fatal("vps setup did not complete properly")
		}
		gitURL, err := local.GetRepoRemote("origin")
		if err != nil {
			log.Fatal(err)
		}

		// Bootstrap remote
		fmt.Printf("Initializing Inertia daemon at %s...\n", inertia.RemoteVPS.IP)
		err = inertia.BootstrapRemote(common.ExtractRepository(common.GetSSHRemoteURL(gitURL)))
		if err != nil {
			log.Fatal(err)
		}
	},
}
