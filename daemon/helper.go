package daemon

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
	docker "github.com/docker/docker/client"
	"github.com/google/go-github/github"
	git "gopkg.in/src-d/go-git.v4"

	"github.com/ubclaunchpad/inertia/common"
)

var (
	// DefaultPort defines the standard daemon port
	// TODO: Reference daemon pkg for this information?
	// We only want the package dependencies to go in one
	// direction, so best to think about how to do this.
	// Clearly cannot ask for this information over HTTP.
	DefaultPort = "8081"

	daemonGithubKeyLocation = "/app/host/.ssh/id_rsa_inertia_deploy"
)

// processPushEvent prints information about the given PushEvent.
func processPushEvent(event *github.PushEvent) {
	repo := event.GetRepo()
	println("Received PushEvent")
	println(fmt.Sprintf("Repository Name: %s", *repo.Name))
	println(fmt.Sprintf("Repository Git URL: %s", *repo.GitURL))
	println(fmt.Sprintf("Ref: %s", event.GetRef()))

	// Clone repository if not available, otherwise skip this step and
	// let deploy() handle the pull.
	err := common.CheckForGit(projectDirectory)
	if err != nil {
		println("No git repository present.")
		err = setUpProject(common.GetSSHRemoteURL(*repo.GitURL), event.GetBaseRef(), os.Stdout)
		if err != nil {
			return
		}
	}

	localRepo, err := git.PlainOpen(projectDirectory)
	if err != nil {
		println(err)
		return
	}

	// Check for matching remotes
	err = common.CompareRemotes(localRepo, common.GetSSHRemoteURL(*repo.GitURL))
	if err != nil {
		println(err)
		return
	}

	// If branches match, deploy
	head, err := localRepo.Head()
	if err != nil {
		println(err)
		return
	}
	if head.Name().Short() == event.GetBaseRef() {
		cli, err := docker.NewEnvClient()
		if err != nil {
			println(err)
			return
		}
		defer cli.Close()
		err = deploy(localRepo, event.GetBaseRef(), cli, os.Stdout)
		if err != nil {
			println(err)
		}
	}
}

// processPullRequestEvent prints information about the given PullRequestEvent.
// Handling PRs is unnecessary because merging one will trigger a PushEvent.
// For now, simply logs events - may in the future do something configured
// by the user.
func processPullRequestEvent(event *github.PullRequestEvent) {
	repo := event.GetRepo()
	pr := event.GetPullRequest()
	merged := "false"
	if *pr.Merged {
		merged = "true"
	}
	println("Received PullRequestEvent")
	println(fmt.Sprintf("Repository Name: %s", *repo.Name))
	println(fmt.Sprintf("Repository Git URL: %s", *repo.GitURL))
	println(fmt.Sprintf("Ref: %s", pr.GetBase().GetRef()))
	println(fmt.Sprintf("Merge status: %v", merged))
}

// GetAPIPrivateKey returns the private RSA key to authenticate HTTP
// requests sent to the daemon. For now, we simply use the GitHub
// deploy key.
func GetAPIPrivateKey(*jwt.Token) (interface{}, error) {
	pemFile, err := os.Open(daemonGithubKeyLocation)
	if err != nil {
		return nil, err
	}
	key, err := common.GetGithubKey(pemFile)
	if err != nil {
		return nil, err
	}
	return []byte(key.String()), nil
}

// setUpProject sets up a project for the first time
func setUpProject(remoteURL, branch string, w io.Writer) error {
	fmt.Fprintln(w, "Setting up project...")
	pemFile, err := os.Open(daemonGithubKeyLocation)
	if err != nil {
		return err
	}
	auth, err := common.GetGithubKey(pemFile)
	if err != nil {
		return err
	}

	// Clone project
	_, err = common.Clone(projectDirectory, remoteURL, branch, auth, w)
	if err != nil {
		if err == common.ErrInvalidGitAuthentication {
			return gitAuthFailedErr(daemonGithubKeyLocation)
		}
		return err
	}
	return nil
}

// gitAuthFailedErr includes the daemon key in the error message
func gitAuthFailedErr(keyloc string) error {
	bytes, err := ioutil.ReadFile(keyloc + ".pub")
	if err != nil {
		bytes = []byte(err.Error() + "\nError reading key - try running 'inertia [REMOTE] init' again: ")
	}
	return errors.New("Access to project repository rejected; did you forget to add\nInertia's deploy key to your repository settings?\n" + string(bytes[:]))
}
