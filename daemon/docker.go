package daemon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
)

// deploy does git pull, docker-compose build, docker-compose up
func deploy(repo *git.Repository, cli *docker.Client, out io.Writer) error {
	pemFile, err := os.Open(daemonGithubKeyLocation)
	if err != nil {
		return err
	}
	auth, err := common.GetGithubKey(pemFile)
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Updating repository...")
	// Pull from working branch
	tree, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = tree.Pull(&git.PullOptions{
		Auth:     auth,
		Depth:    2,
		Progress: out,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		if err == transport.ErrInvalidAuthMethod || err == transport.ErrAuthorizationFailed || strings.Contains(err.Error(), "unable to authenticate") {
			bytes, err := ioutil.ReadFile(daemonGithubKeyLocation + ".pub")
			if err != nil {
				bytes = []byte("Error reading key - try running 'inertia [REMOTE] init' again.")
			}
			return errors.New("Access to project repository rejected; did you forget to add\nInertia's deploy key to your repository settings?\n" + string(bytes[:]))
		} else if err == git.ErrForceNeeded {
			// If pull fails, attempt a force pull before returning error
			fmt.Fprint(out, "Force pull required - making a fresh clone...")
			_, err := common.ForcePull(projectDirectory, repo, auth, out)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Kill active project containers if there are any
	fmt.Fprintln(out, "Shutting down active containers...")
	err = killActiveContainers(cli, out)
	if err != nil {
		return err
	}

	// Build and run project - the following code performs the bash
	// equivalent of:
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
	fmt.Fprintln(out, "Building project...")
	ctx := context.Background()
	resp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image:      dockerCompose,
			WorkingDir: "/build/project",
			Env:        []string{"HOME=/build"},
			Cmd: []string{
				"up",
				"--build",
			},
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
		killErr := killActiveContainers(cli, out)
		if killErr != nil {
			fmt.Fprintln(out, err)
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
func killActiveContainers(cli *docker.Client, out io.Writer) error {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	for _, container := range containers {
		if container.Names[0] != "/inertia-daemon" {
			fmt.Fprintln(out, "Killing "+container.Image+" ("+container.Names[0]+")...")
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
	if len(report.ContainersDeleted) > 0 {
		fmt.Fprintln(out, "Removed "+strings.Join(report.ContainersDeleted, ", "))
	}
	return nil
}
