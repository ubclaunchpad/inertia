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
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
)

var (
	// specify location of deployed project
	projectDirectory = "/app/host/project"

	// specify docker-compose version
	dockerCompose = "docker/compose:1.18.0"

	// specify common responses here
	okResp           = "I'm a little Webhook, short and stout!"
	noContainersResp = "There are currently no active containers."

	daemonGithubKeyLocation = "/app/host/.ssh/id_rsa_inertia_deploy"
	defaultSecret           = "inertia"
)

/*
 * CLI Commands
 */

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
		cli, err := docker.NewEnvClient()
		if err != nil {
			log.WithError(err)
			println("Failed to pull docker-compose image - shutting down daemon.")
			return
		}
		_, err = cli.ImagePull(context.Background(), dockerCompose, types.ImagePullOptions{})
		if err != nil {
			log.WithError(err)
			println("Failed to pull docker-compose image - shutting down daemon.")
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
		// Example usage of `authorized' decorator.
		mux.HandleFunc("/health-check", authorized(healthCheckHandler, getAPIPrivateKey))
		mux.HandleFunc("/", gitHubWebHookHandler)
		mux.HandleFunc("/up", authorized(upHandler, getAPIPrivateKey))
		mux.HandleFunc("/down", authorized(downHandler, getAPIPrivateKey))
		mux.HandleFunc("/status", authorized(statusHandler, getAPIPrivateKey))
		mux.HandleFunc("/reset", authorized(resetHandler, getAPIPrivateKey))
		log.Fatal(http.ListenAndServe(":"+port, mux))
	},
}

// tokenCmd represents the daemon run command
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Produce an API token to use with the daemon",
	Long: `Produce an API token to use with the daemon,
	Created using an RSA private key.`,
	Run: func(cmd *cobra.Command, args []string) {
		keyBytes, err := getAPIPrivateKey(nil)
		if err != nil {
			log.Fatal(err)
		}

		token, err := generateToken(keyBytes.([]byte))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(token)
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(runCmd)
	daemonCmd.AddCommand(tokenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// daemonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	runCmd.Flags().StringP("port", "p", "8081", "Set port for daemon to run on")
}

/*
 * Handlers
 */

// healthCheckHandler returns a 200 if the daemon is happy.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, okResp)
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

// upHandler tries to bring the deployment online
func upHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("UP request received")

	// Get github URL from up request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var upReq UpRequest
	err = json.Unmarshal(body, &upReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	remoteURL := upReq.Repo

	// Check for existing git repository, clone if no git repository exists.
	err = checkForGit(projectDirectory)
	if err != nil {
		log.Println("No git repository present - cloning from POST event...")
		pemFile, err := os.Open(daemonGithubKeyLocation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
			return
		}
		auth, err := getGithubKey(pemFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Clone project
		log.Println("Attempting to clone " + remoteURL)
		_, err = clone(projectDirectory, remoteURL, auth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
			err = removeContents(projectDirectory)
			if err != nil {
				log.WithError(err)
			}
			return
		}
	}

	repo, err := git.PlainOpen(projectDirectory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check for matching remotes
	err = compareRemotes(repo, remoteURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		return
	}

	// Update and deploy project
	cli, err := docker.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	err = deploy(repo, cli)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Project up and running!")
}

// downHandler tries to take the deployment offline
func downHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DOWN request received")

	cli, err := docker.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	// Error if no project containers are active, but try to kill
	// everything anyway in case the docker-compose image is still
	// active
	_, err = getActiveContainers(cli)
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		err = killActiveContainers(cli)
		if err != nil {
			log.WithError(err)
		}
		return
	}

	err = killActiveContainers(cli)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Project shut down.")
}

// statusHandler lists currently active project containers
func statusHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("STATUS request received")

	// Get status of repository
	repo, err := git.PlainOpen(projectDirectory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		return
	}
	remotes, err := repo.Remotes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		return
	}
	remoteURL := remotes[0].Config().URLs[0]
	head, err := repo.Head()
	if err != nil {
		http.Error(w, err.Error(), http.StatusPreconditionFailed)
		return
	}
	repoStatus := remoteURL + "\n" + head.String() + "\n"

	// Get containers
	cli, err := docker.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	containers, err := getActiveContainers(cli)
	if err != nil {
		if err.Error() == noContainersResp {
			// This is different from having 2 containers active -
			// noContainersResp means that no attempt to build the project
			// was made or the project was cleanly shut down.
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, repoStatus+noContainersResp)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If there are only 2 containers active, that means that a build
	// attempt was made but only the daemon and the docker-compose containers
	// are active, indicating a build failure.
	if len(containers) == 2 {
		errorString := repoStatus + "It appears that an attempt to start your project was made but the build failed."
		http.Error(w, errorString, http.StatusNotFound)
		return
	}

	ignore := map[string]bool{
		"/inertia-daemon": true,
		"/docker-compose": true,
	}
	// Only list project containers
	activeContainers := "Active containers:"
	for _, container := range containers {
		if !ignore[container.Names[0]] {
			activeContainers += "\n" + container.Image + " (" + container.Names[0] + ")"
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, repoStatus+activeContainers)
}

// resetHandler shuts down and wipes the project directory
func resetHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("RESET request received")

	cli, err := docker.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()
	err = killActiveContainers(cli)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = removeContents(projectDirectory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Project removed from remote.")
}

/*
 * Helper Functions
 */

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
		pemFile, err := os.Open(daemonGithubKeyLocation)
		if err != nil {
			log.Println("No GitHub key found: " + err.Error())
			return
		}
		auth, err := getGithubKey(pemFile)
		if err != nil {
			log.Println("Github key couldn't be read: " + err.Error())
			return
		}
		_, err = clone(projectDirectory, getSSHRemoteURL(*repo.GitURL), auth)
		if err != nil {
			log.Println("Clone failed: " + err.Error())
			err = removeContents(projectDirectory)
			if err != nil {
				log.WithError(err)
			}
			return
		}
	}

	localRepo, err := git.PlainOpen(projectDirectory)
	if err != nil {
		log.WithError(err)
		return
	}

	// Check for matching remotes
	err = compareRemotes(localRepo, getSSHRemoteURL(*repo.GitURL))
	if err != nil {
		log.WithError(err)
		return
	}

	// Deploy project
	cli, err := docker.NewEnvClient()
	if err != nil {
		log.WithError(err)
		return
	}
	defer cli.Close()
	err = deploy(localRepo, cli)
	if err != nil {
		log.WithError(err)
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
	log.Println("Received PullRequestEvent")
	log.Println(fmt.Sprintf("Repository Name: %s", *repo.Name))
	log.Println(fmt.Sprintf("Repository Git URL: %s", *repo.GitURL))
	log.Println(fmt.Sprintf("Ref: %s", pr.GetBase().GetRef()))
	log.Println(fmt.Sprintf("Merge status: %v", merged))
}

// deploy does git pull, docker-compose build, docker-compose up
func deploy(repo *git.Repository, cli *docker.Client) error {
	pemFile, err := os.Open(daemonGithubKeyLocation)
	if err != nil {
		return err
	}
	auth, err := getGithubKey(pemFile)
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
	// This starts a new container running a docker-compose image for
	// the sole purpose of building the project. This container is
	// separate from the daemon and the user's project, and is the
	// second container to require access to the docker socket.
	// See https://cloud.google.com/community/tutorials/docker-compose-on-container-optimized-os
	log.Println("Bringing project online.")
	ctx := context.Background()
	resp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image:      dockerCompose,
			WorkingDir: "/build/project",
			Env:        []string{"HOME:/build"},
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

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// Check if build failed abruptly
	time.Sleep(2 * time.Second)
	_, err = getActiveContainers(cli)
	if err != nil {
		err := killActiveContainers(cli)
		return errors.New("Docker-compose failed: " + err.Error())
	}

	return nil
}

// getActiveContainers returns all active containers and returns and error
// if the Daemon is the only active container
func getActiveContainers(cli *docker.Client) ([]types.Container, error) {
	containers, err := cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	if err != nil {
		return nil, err
	}

	// Error if only one container (daemon) is active
	if len(containers) <= 1 {
		return nil, errors.New(noContainersResp)
	}

	return containers, nil
}

// killActiveContainers kills all active project containers (ie not including daemon)
func killActiveContainers(cli *docker.Client) error {
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
