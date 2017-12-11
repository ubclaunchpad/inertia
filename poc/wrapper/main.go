// POC for running RPC over SSH.
// Try the following:
//    go run main.go "git clone <my_repo>"
// 	  go run main.go "ls -lsa"
// Special command for bootstrapping a server:
//
// go run main.go bootstrap

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"golang.org/x/crypto/ssh"
)

// RemoteVPS holds access to a remote instance.
type RemoteVPS struct {
	User string
	IP   string
	PEM  string
}

func main() {
	remoteCmd := os.Args[1]

	remote := &RemoteVPS{
		User: "brunocodesbad",
		IP:   "35.227.171.49",
		PEM:  "/Users/chadlagore/.ssh/id_inertia",
	}

	if remoteCmd == "bootstrap" {
		println("Bootstrapping VPS, this could take some time, maybe opt for a goroutine next time...")
		result, err := remote.RunSSHScript("bootstrap.sh")

		if err != nil {
			fmt.Println(err)
			fmt.Println(result)
			return
		}

		fmt.Println(result)

	} else {
		result, err := remote.RunSSHCommand(remoteCmd)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(result)
	}
}

// GetHost creates the user@IP string.
func (remote *RemoteVPS) GetHost() string {
	return remote.User + "@" + remote.IP
}

// RunSSHCommand runs a command remotely.
func (remote *RemoteVPS) RunSSHCommand(remoteCmd string) (*bytes.Buffer, error) {
	cmd := exec.Command("ssh", "-i", remote.PEM, "-t", remote.GetHost(), remoteCmd)

	// Capture result.
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		out = stderr
	}

	return &out, err
}

// RunSSHScript runs a command remotely.
func (remote *RemoteVPS) RunSSHScript(localScript string) (*bytes.Buffer, error) {
	// To run this, we actually have to copy the script to the remote server
	// using scp.
	clientConfig, _ := auth.PrivateKey(
		remote.User, remote.PEM, ssh.InsecureIgnoreHostKey())

	client := scp.NewClient(remote.IP+":22", &clientConfig)

	// Connect to the remote server
	err := client.Connect()
	if err != nil {
		fmt.Println("Couldn't establisch a connection to the remote server ", err)
		return nil, err
	}

	// Open a file
	f, _ := os.Open(localScript)

	// Close session after the file has been copied
	defer client.Session.Close()

	// Close the file after it has been copied
	defer f.Close()

	// Finaly, copy the file over
	// Usage: CopyFile(fileReader, remotePath, permission)
	filename := filepath.Base(localScript)
	remoteLoc := "/tmp/" + filename
	client.CopyFile(f, remoteLoc, "0644")

	// Run the script remotely.
	return remote.RunSSHCommand("sh " + remoteLoc)

	// TODO: Clean up server.
}
