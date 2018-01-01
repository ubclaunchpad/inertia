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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	// specify location of deployed project
	projectDirectory = "/app/host/project"

	// specify docker-compose version
	dockerCompose = "docker/compose:1.18.0"

	defaultSecret = "inertia"

	// specify common responses here
	okResp           = "I'm a little Webhook, short and stout!"
	noContainersResp = "There are currently no active containers."
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
		// Download docker-compose image
		println("Downloading docker-compose...")
		cli, err := client.NewEnvClient()
		if err != nil {
			return
		}
		_, err = cli.ImagePull(context.Background(), dockerCompose, types.ImagePullOptions{})
		if err != nil {
			cli.Close()
			return
		}
		cli.Close()

		// Run daemon on port
		port, err := cmd.Flags().GetString("port")
		if err != nil {
			log.WithError(err)
		}
		println("Serving daemon on port " + port)
		mux := http.NewServeMux()
		mux.HandleFunc("/", gitHubWebHookHandler)
		mux.HandleFunc("/up", upHandler)
		mux.HandleFunc("/down", downHandler)
		mux.HandleFunc("/status", statusHandler)
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

	// Clone repository if not available
	err := checkForGit(projectDirectory)
	if err != nil {
		log.Println("No git repository present - cloning from push event...")
		auth, err := getGithubKey()
		if err != nil {
			return
		}
		_, err = git.PlainClone(projectDirectory, false, &git.CloneOptions{
			URL:  getSSHRemoteURL(*repo.GitURL),
			Auth: auth,
		})
		if err != nil {
			log.Println("Clone failed: " + err.Error())
			removeContents(projectDirectory)
			return
		}
	}

	localRepo, err := git.PlainOpen(projectDirectory)
	if err != nil {
		log.Println(err.Error())
	}

	// Deploy project
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Println(err.Error())
	}
	defer cli.Close()

	err = deploy(localRepo, cli)
	if err != nil {
		log.Println(err.Error())
	}
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
	cli, err := client.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cli.Close()

	// Check for existing git repository, clone if no git repository exists.
	err = checkForGit(projectDirectory)
	if err != nil {
		log.Println("No git repository present - cloning from POST event...")
		auth, err := getGithubKey()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Get github URL from up request for cloning
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer r.Body.Close()
		var upReq UpRequest
		err = json.Unmarshal(body, &upReq)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Clone project
		remoteURL := upReq.Repo
		log.Println("Attempting to clone " + remoteURL)
		_, err = git.PlainClone(projectDirectory, false, &git.CloneOptions{
			URL:  remoteURL,
			Auth: auth,
		})
		if err != nil {
			removeContents(projectDirectory)
			http.Error(w, err.Error(), 500)
			return
		}
	}

	// Update and deploy the repository
	repo, err := git.PlainOpen(projectDirectory)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = deploy(repo, cli)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Project up and running!")
}

// deploy does git pull, docker-compose build, docker-compose up
func deploy(repo *git.Repository, cli *client.Client) error {
	auth, err := getGithubKey()
	if err != nil {
		return err
	}

	// Pull from working branch
	tree, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = tree.Pull(&git.PullOptions{
		Auth: auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		// If pull fails, attempt a force pull before returning error
		log.Println("Pull failed - attempting a fresh clone...")
		repo, err = forcePull(repo, auth)
		if err != nil {
			return err
		}
	}

	// Kill active project containers if there are any
	err = killActiveContainers(cli)
	if err != nil {
		return err
	}

	// Build and run project - the following code performs the
	// shell equivalent of:
	//
	//    docker run -d \
	// 	    -v /var/run/docker.sock:/var/run/docker.sock \
	// 	    -v $HOME:/build \
	// 	    -w="/build/project" \
	// 	    docker/compose:1.18.0 up --build
	//
	// This runs a docker-compose image to build the Docker project.
	// See https://cloud.google.com/community/tutorials/docker-compose-on-container-optimized-os
	log.Println("Bringing project online.")
	ctx := context.Background()
	resp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image:      dockerCompose,
			WorkingDir: "/build/project",
			Cmd:        []string{"up", "--build"},
		},
		&container.HostConfig{
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
				os.Getenv("HOME") + ":/build",
			},
		}, nil, "docker-compose",
	)
	if err != nil {
		return err
	}
	if len(resp.Warnings) > 0 {
		warnings := strings.Join(resp.Warnings, "\n")
		return errors.New(warnings)
	}

	return cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
}

// downHandler tries to take the deployment offline
func downHandler(w http.ResponseWriter, r *http.Request) {
	cli, err := client.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cli.Close()

	err = killActiveContainers(cli)
	if err != nil {
		http.Error(w, err.Error(), 412)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Project shut down.")
}

// statusHandler lists currently active project containers
func statusHandler(w http.ResponseWriter, r *http.Request) {
	cli, err := client.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cli.Close()
	containers, err := getActiveContainers(cli)
	if err != nil {
		if err.Error() == noContainersResp {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, noContainersResp)
			return
		}
		http.Error(w, err.Error(), 500)
		return
	}

	ignore := map[string]bool{
		"/inertia-daemon": true,
		"/docker-compose": true,
	}

	activeContainers := "Active containers:"
	for _, container := range containers {
		if !ignore[container.Names[0]] {
			activeContainers += "\n" + container.Image + " (" + container.Names[0] + ")"
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, activeContainers)
}

// getGithubKey returns the key generated by 'inertia remote bootstrap [REMOTE]'
func getGithubKey() (ssh.AuthMethod, error) {
	pemFile := "/app/host/.ssh/id_rsa_inertia_deploy"
	return ssh.NewPublicKeysFromFile("git", pemFile, "")
}

// getActiveContainers returns all active containers and returns and error
// if the Daemon is the only active container
func getActiveContainers(cli *client.Client) ([]types.Container, error) {
	containers, err := cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	if err != nil {
		return nil, err
	}

	// If 2 or fewer containers are active, that means either
	// only the daemon is active or only the daemon and the
	// docker-compose image is active
	if len(containers) <= 2 {
		return nil, errors.New(noContainersResp)
	}

	return containers, nil
}

// killActiveContainers kills all active project containers (ie not including daemon)
func killActiveContainers(cli *client.Client) error {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	for _, container := range containers {
		if container.Names[0] != "/inertia-daemon" {
			log.Println("Killing " + container.Image + " (" + container.Names[0] + ")...")
			err := cli.ContainerKill(ctx, container.ID, "SIGKILL")
			if err != nil {
				return err
			}
			err = cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// forcePull deletes the project directory and makes a fresh clone of given repo
// git.Worktree.Pull() only supports merges that can be resolved as a fast-forward
func forcePull(repo *git.Repository, auth ssh.AuthMethod) (*git.Repository, error) {
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}
	remoteURL := getSSHRemoteURL(remotes[0].Config().URLs[0])
	err = removeContents(projectDirectory)
	if err != nil {
		repo, err = git.PlainClone(projectDirectory, false, &git.CloneOptions{
			URL:  remoteURL,
			Auth: auth,
		})
		if err != nil {
			removeContents(projectDirectory)
			return nil, err
		}
	}
	return repo, nil
}
