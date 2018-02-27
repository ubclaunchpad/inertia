package main

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
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
