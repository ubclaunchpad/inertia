// Proof-of-concept for programmatically generating a deploy key to register
// with GitHub. To test, run go main.go github.com/<yourname>/<yourrepo>.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

var (
	sshDir = "/Users/chadlagore/.ssh/"
)

func main() {
	repo := os.Args[1]
	CreateDeployKey(repo)
}

// CreateDeployKey creates and saves deploy keys to disk, prompts
// user with the public key (so they can copy paste to GitHub).
// Alternatively, this function can push the deploy key to GitHub,
// but we need user logon.
func CreateDeployKey(repo string) {
	repoName := filepath.Base(repo) + "-inertia"

	pemFile := filepath.Join(sshDir, repoName)
	pubFile := filepath.Join(sshDir, repoName+".pub")

	fmt.Println("New PEM file: " + pemFile)
	fmt.Println("New PUB file: " + pubFile)

	err := SSHKeyGen(pubFile, pemFile, 2014)
	if err != nil {
		fmt.Println(err)
		return
	}

	pub, err := os.Open(pubFile)
	defer pub.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	deployKeyURL := "https://" + repo + "/settings/keys/new"
	fmt.Println("\n😏  Your new inertia public deploy key:\n")

	io.Copy(os.Stdout, pub)

	fmt.Println("\nAdd it here: " + deployKeyURL)
}

// SSHKeyGen creates a public-private key pair much like ssh-keygen
// command line utility, except it does not prompt for pw. Writes
// result to disk.
// https://stackoverflow.com/a/34347463.
func SSHKeyGen(pubLoc, pemLoc string, size int) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return err
	}

	// Create private key file at pemLoc.
	privateKeyFile, err := os.Create(pemLoc)
	defer privateKeyFile.Close()
	if err != nil {
		return err
	}

	// Use PEM encoding.
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return err
	}

	// Create a public key from private.
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	// Write to public file locatio.
	return ioutil.WriteFile(pubLoc, ssh.MarshalAuthorizedKey(pub), 0655)
}
