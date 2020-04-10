package main

import (
	"fmt"
	"os"

	"github.com/ubclaunchpad/inertia/cmd"
	"github.com/ubclaunchpad/inertia/local"
)

// Version indicates the current version of Inertia
var Version string

func main() {
	if err := cmd.NewInertiaCmd(Version, local.InertiaDir(), true).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
