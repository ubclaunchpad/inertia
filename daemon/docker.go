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

package daemon

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/ubclaunchpad/inertia/common"
	git "gopkg.in/src-d/go-git.v4"
)

// deploy does git pull, docker-compose build, docker-compose up
func deploy(repo *git.Repository, cli *docker.Client) error {
	pemFile, err := os.Open(daemonGithubKeyLocation)
	if err != nil {
		return err
	}
	auth, err := common.GetGithubKey(pemFile)
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
		_, err = common.ForcePull(projectDirectory, repo, auth)
		if err != nil {
			return err
		}

		// Wait arbitrary amount of time for clone to complete
		// TODO: find a better way to do this
		time.Sleep(2 * time.Second)
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
	time.Sleep(3 * time.Second)
	_, err = getActiveContainers(cli)
	if err != nil {
		killErr := killActiveContainers(cli)
		if killErr != nil {
			log.WithError(err)
		}
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
		}
	}

	report, err := cli.ContainersPrune(ctx, filters.Args{})
	if err != nil {
		return err
	}
	log.Println("Removed " + strings.Join(report.ContainersDeleted, ", "))
	return nil
}
