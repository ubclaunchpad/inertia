// Proof-of-concept for programmatically cloning a GitHub repo over SSH using
// a preset deploy key. To test, add a deploy key to the Inertia repo, set the
// pemFile variable to point to that key's private key file, and run.
package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	url     = "git@github.com:ubclaunchpad/inertia.git"
	pemFile = "/Users/jordan/.ssh/id_rsa"
)

func main() {
	dir, err := ioutil.TempDir("", "poc-ssh")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(dir)

	auth, err := ssh.NewPublicKeysFromFile("git", pemFile, "")
	if err != nil {
		log.Fatal(err)
	}

	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:  url,
		Auth: auth,
	})
	if err != nil {
		log.Fatal(err)
	}

	ref, err := r.Head()
	if err != nil {
		log.Fatal(err)
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		log.Fatal(err)
	}

	log.Println(commit)
}
