// Copyright © 2017 UBC Launch Pad team@ubclaunchpad.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	defaultSecret = "inertia"
	okResp        = "I'm a little Webhook, short and stout!"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Configure daemon behaviour from command line",
	Args:  cobra.MinimumNArgs(1),
	Run:   func(cmd *cobra.Command, args []string) {},
}

// runCmd represents the daemon run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the daemon",
	Long: `Run the daemon on a port.
Example:

inertia daemon run -p 8081`,
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetString("port")
		if err != nil {
			log.WithError(err)
		}
		println("Serving daemon on port " + port)
		mux := http.NewServeMux()
		mux.HandleFunc("/", gitHubWebHookHandler)
		mux.HandleFunc("/up", upHandler)
		mux.HandleFunc("/down", downHandler)
		log.Fatal(http.ListenAndServe(":"+port, mux))
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// daemonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	runCmd.Flags().StringP("port", "p", "8081", "Set port for daemon to run on")
}

// gitHubWebHookHandler writes a response to a request into the given ResponseWriter.
func gitHubWebHookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, okResp)

	payload, err := github.ValidatePayload(r, []byte(defaultSecret))
	if err != nil {
		log.Println(err)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Println(err)
		return
	}

	switch event := event.(type) {
	case *github.PushEvent:
		processPushEvent(event)
	case *github.PullRequestEvent:
		processPullRequestEvent(event)
	default:
		log.Println("Unrecognized event type")
	}
}

// processPushEvent prints information about the given PushEvent.
func processPushEvent(event *github.PushEvent) {
	repo := event.GetRepo()
	log.Println("Received PushEvent")
	log.Println(fmt.Sprintf("Repository Name: %s", *repo.Name))
	log.Println(fmt.Sprintf("Repository Git URL: %s", *repo.GitURL))
	log.Println(fmt.Sprintf("Ref: %s", event.GetRef()))

	// Deploy
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err.Error())
	}
	localRepo, err := git.PlainOpen(filepath.Join(cwd, ".git"))
	if err != nil {
		log.Println(err.Error())
	}
	deploy(localRepo)
}

// processPullREquestEvent prints information about the given PullRequestEvent.
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
	log.Println("Received PullRequestEvent")
	log.Println(fmt.Sprintf("Repository Name: %s", *repo.Name))
	log.Println(fmt.Sprintf("Repository Git URL: %s", *repo.GitURL))
	log.Println(fmt.Sprintf("Ref: %s", pr.GetBase().GetRef()))
	log.Println(fmt.Sprintf("Merge status: %v", merged))
}

// upHandler tries to bring the deployment online
func upHandler(w http.ResponseWriter, r *http.Request) {
	cwd, err := os.Getwd()
	if err != nil {
		http.Error(w, err.Error(), 501)
	}
	// TODD: Check for repo.
	// If no repo,
	// 1. clone
	// 2. build
	// 3. run.
	// If repo,
	// Check for existing containers.
	// Check for existing images.

	err = CheckForGit()
	if err != nil {
		// Clone project
		remoteURL := "" // TODO
		cfg, err := GetProjectConfigFromDisk()
		if err != nil {
			log.Println(err.Error())
		}
		pemFile := cfg.CurrentRemoteVPS.PEM
		auth, err := ssh.NewPublicKeysFromFile("git", pemFile, "")
		if err != nil {
			log.Println(err.Error())
		}
		repo, err := git.PlainClone(cwd, false, &git.CloneOptions{
			URL:  remoteURL,
			Auth: auth,
		})
		if err != nil {
			http.Error(w, err.Error(), 501)
		}
		err = deploy(repo)
		if err != nil {
			http.Error(w, err.Error(), 501)
		}
	} else {
		// Pull project's current branch
		repo, err := git.PlainOpen(filepath.Join(cwd, ".git"))
		if err != nil {
			http.Error(w, err.Error(), 501)
		}
		err = deploy(repo)
		if err != nil {
			http.Error(w, err.Error(), 501)
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Project up and running!")
}

// deploy does git pull, docker-compose build, docker-compose up
func deploy(repo *git.Repository) error {
	cfg, err := GetProjectConfigFromDisk()
	if err != nil {
		log.Println(err.Error())
	}
	pemFile := cfg.CurrentRemoteVPS.PEM
	auth, err := ssh.NewPublicKeysFromFile("git", pemFile, "")
	if err != nil {
		log.Println(err.Error())
	}

	tree, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = tree.Pull(&git.PullOptions{
		Auth: auth,
	})

	// Build and run
	daemonCmd, err := Asset("cmd/bootstrap/docker.sh")
	if err != nil {
		return err
	}
	cmd := exec.Command(string(daemonCmd))
	return cmd.Run()
}

// downHandler tries to take the deployment offline
func downHandler(w http.ResponseWriter, r *http.Request) {
	// Check if daemon is the only active container
	cli, err := client.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), 501)
	}
	containers, err := cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	if err != nil {
		http.Error(w, err.Error(), 501)
	}
	if len(containers) == 1 {
		http.Error(w, "No Docker containers are currently active", 501)
	}

	// Take project offline
	daemonCmd, err := Asset("cmd/bootstrap/project-down.sh")
	if err != nil {
		http.Error(w, err.Error(), 501)
	}
	cmd := exec.Command(string(daemonCmd))
	err = cmd.Run()
	if err != nil {
		http.Error(w, err.Error(), 501)
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Project shut down.")
}
