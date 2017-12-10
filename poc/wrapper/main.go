package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// RemoteVPS holds access to a remote instance.
type RemoteVPS struct {
	Host string
	PEM  string
}

func main() {
	remoteCmd := os.Args[1]

	remote := &RemoteVPS{
		Host: "brunocodesbad@35.227.171.49",
		PEM:  "/Users/chadlagore/.ssh/id_inertia",
	}

	result, err := remote.RunSSHCommand(remoteCmd)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(result)
}

// RunSSHCommand runs a command remotely.
func (remote *RemoteVPS) RunSSHCommand(remoteCmd string) (*bytes.Buffer, error) {
	cmd := exec.Command(
		"ssh", "-i", remote.PEM,
		"-t", remote.Host, remoteCmd,
	)

	// Capture result.
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	return &out, err
}
