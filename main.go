package main

import (
	"fmt"
	"os"

	"github.com/ubclaunchpad/inertia/cmd"
)

// Version indicates the current version of Inertia
var Version string

func main() {
	if err := cmd.NewInertiaCmd(Version).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
