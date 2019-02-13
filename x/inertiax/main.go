package main

import (
	"fmt"
	"os"

	"github.com/ubclaunchpad/inertia/cmd"
)

// Version denotes the version of the binary
var Version string

func main() {
	var root = cmd.NewInertiaCmd(Version)
	AttachXCmds(root)
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
