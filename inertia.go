package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cmd"
)

var (
	// Version is the current build of Inertia
	Version string
)

func setVersion(c *cobra.Command) {
	if Version != "" {
		c.Version = Version
	} else {
		c.Version = "latest"
	}
}

func main() {
	root := cmd.Root
	setVersion(root)
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
