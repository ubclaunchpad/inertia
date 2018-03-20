package project

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	// Directory specifies the location of deployed project
	Directory = "/app/host/project"

	// DockerComposeVersion is the docker-compose version used by the daemon
	DockerComposeVersion = "docker/compose:1.18.0"

	// NoContainersResp is the response to indicate that no containers are active
	NoContainersResp = "There are currently no active containers."
)

// Deploy does git pull, docker-compose build, docker-compose up
func Deploy(auth ssh.AuthMethod, repo *git.Repository, branch string, cli *docker.Client, out io.Writer) error {
	fmt.Println(out, "Deploying repository...")

	// Pull from given branch and check out if needed
	err := common.UpdateRepository(Directory, repo, branch, auth, out)
	if err != nil {
		return err
	}

	// Kill active project containers if there are any
	err = KillActiveContainers(cli, out)
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
	fmt.Fprintln(out, "Setting up docker-compose...")
	ctx := context.Background()
	resp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image:      DockerComposeVersion,
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

	fmt.Fprintln(out, "Building project...")
	return cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})

	// Check if build failed abruptly
	// This is disabled until a more consistent way of detecting build
	// failures is implemented.
	/*
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
	*/
}

// GetActiveContainers returns all active containers and returns and error
// if the Daemon is the only active container
func GetActiveContainers(cli *docker.Client) ([]types.Container, error) {
	containers, err := cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	if err != nil {
		return nil, err
	}

	// Error if only one container (daemon) is active
	if len(containers) <= 1 {
		return nil, errors.New(NoContainersResp)
	}

	return containers, nil
}

// KillActiveContainers kills all active project containers (ie not including daemon)
func KillActiveContainers(cli *docker.Client, out io.Writer) error {
	fmt.Fprintln(out, "Shutting down active containers...")
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	for _, container := range containers {
		if container.Names[0] != "/inertia-daemon" {
			fmt.Fprintln(out, "Killing "+container.Image+" ("+container.Names[0]+")...")
			timeout := 10 * time.Second
			err := cli.ContainerStop(ctx, container.ID, &timeout)
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
