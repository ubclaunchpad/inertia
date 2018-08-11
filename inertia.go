package main

import (
	"fmt"
	"os"

	"github.com/ubclaunchpad/inertia/cmd"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
