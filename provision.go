package main

// initCmd represents the init command
import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
	"github.com/ubclaunchpad/inertia/provision"
)

// Initialize "inertia" commands regarding basic configuration
func init() {
	cmdProvision.AddCommand(cmdProvisionECS)
	cmdRoot.AddCommand(cmdProvision)

	cmdProvision.Flags().String("version", Version, "Specify Inertia daemon version to use")
}

var cmdProvision = &cobra.Command{
	Use:   "provision",
	Short: "Provision a new VPS set up for Inertia",
	Long:  `Provision a new VPS instance set up for continuous deployment with Inertia.`,
}

var cmdProvisionECS = &cobra.Command{
	Use:   "ecs [name]",
	Short: "Provision a new Amazon ECS instance",
	Long: `Provision a new Amazon ECS instance and set it up for continuous deployment
	with Inertia.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure project initialized.
		config, path, err := local.GetProjectConfigFromDisk()
		if err != nil {
			log.Fatal(err)
		}

		// Create VPS instance
		prov := provision.NewEC2Provisioner()
		// todo: allow config
		address, err := prov.CreateInstance("m3.medium", "us-east-1b")
		if err != nil {
			log.Fatal(err)
		}

		// Save new remote
		branch, err := local.GetRepoCurrentBranch()
		if err != nil {
			log.Fatal(err)
		}
		config.AddRemote(&cfg.RemoteVPS{
			Name:    args[0],
			IP:      address,
			User:    "", // todo
			PEM:     "", // todo
			Branch:  branch,
			SSHPort: "22",
		})
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
