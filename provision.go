package main

// initCmd represents the init command
import (
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
		"type", "t", "m3.medium", "The ec2 instance type to instantiate",
	)
	cmdProvisionECS.Flags().Bool(
		"from-env", false, "Load ec2 credentials from ENV",
	)
	cmdProvision.AddCommand(cmdProvisionECS)
	cmdRoot.AddCommand(cmdProvision)
}

var cmdProvision = &cobra.Command{
	Use:   "provision",
	Short: "Provision a new VPS set up for Inertia",
	Long:  `Provision a new VPS instance set up for continuous deployment with Inertia.`,
}

var cmdProvisionECS = &cobra.Command{
	Use:   "ec2 [name]",
	Short: "Provision a new Amazon EC2 instance",
	Long: `Provision a new Amazon EC2 instance and set it up for continuous deployment
	with Inertia.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, path, err := local.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		// Load flags
		fromEnv, _ := cmd.Flags().GetBool("from-env")
		instanceType, _ := cmd.Flags().GetString("type")

		// Create VPS instance
		var prov *provision.EC2Provisioner
		if fromEnv {
			id, secret, token, err := enterEC2CredentialsWalkthrough(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
			prov = provision.NewEC2Provisioner(id, secret, token)
		} else {
			prov = provision.NewEC2ProvisionerFromEnv()
		}

		// List regions and prompt for input
		regions, err := prov.ListRegions()
		if err != nil {
			log.Fatal(err)
		}
		region, err := chooseFromListWalkthrough(os.Stdin, "region", regions)
		if err != nil {
			log.Fatal(err)
		}

		// List image options and prompt for input
		images, err := prov.ListImageOptions(region)
		if err != nil {
			log.Fatal(err)
		}
		image, err := chooseFromListWalkthrough(os.Stdin, "image", images)
		if err != nil {
			log.Fatal(err)
		}

		// Create instance from input
		remote, err := prov.CreateInstance(args[0], image, instanceType, region)
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

		// Init the new instance
		inertia, found := client.NewClient(args[0], config)
		if !found {
			log.Fatal("vps setup did not complete properly")
		}
		gitURL, err := local.GetRepoRemote("origin")
		if err != nil {
			log.Fatal(err)
		}
		err = inertia.BootstrapRemote(common.ExtractRepository(gitURL))
		if err != nil {
			log.Fatal(err)
		}
	},
}
