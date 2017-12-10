package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

var (
	pemFileLoc = "/Users/chadlagore/.ssh/id_inertia"
	host       = "brunocodesbad@35.227.171.49"
)

func main() {
	remoteCmd := os.Args[1]

	result, err := RunSSHCommand(host, remoteCmd)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(result)
}

// RunSSHCommand runs a command remotely.
func RunSSHCommand(host, remoteCmd string) (*bytes.Buffer, error) {
	cmd := exec.Command(
		"ssh", "-i", pemFileLoc,
		"-t", host, remoteCmd,
	)

	// Capture result.
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	return &out, err
}
