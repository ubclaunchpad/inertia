package project

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
)

// dockerCompose builds and runs project using docker-compose -
// the following code performs the bash equivalent of:
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
func dockerCompose(d *Deployment, cli *docker.Client, out io.Writer) error {
	fmt.Fprintln(out, "Setting up docker-compose...")
	ctx := context.Background()

	resp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image:      DockerComposeVersion,
			WorkingDir: "/build/project",
			Env:        []string{"HOME=/build"},
			Cmd:        []string{"-p", d.project, "build"},
		},
		&container.HostConfig{
			AutoRemove: true,
			Binds: []string{
				os.Getenv("HOME") + "/project/:/build",
				"/var/run/docker.sock:/var/run/docker.sock",
			},
		}, nil, BuildStageName,
	)
	if err != nil {
		return err
	}
	if len(resp.Warnings) > 0 {
		fmt.Fprintln(out, "Warnings encountered on docker-compose build.")
		warnings := strings.Join(resp.Warnings, "\n")
		return errors.New(warnings)
	}

	// Start the herokuish container to build project
	fmt.Fprintln(out, "Building project...")
	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// Attach logs and report build progress until container exits
	reader, err := d.Logs(cli, LogOptions{Container: resp.ID, Stream: true})
	if err != nil {
		return err
	}
	stop := make(chan struct{})
	go common.FlushRoutine(out, reader, stop)
	status, err := cli.ContainerWait(ctx, resp.ID)
	close(stop)
	reader.Close()
	if err != nil {
		return err
	}
	if status != 0 {
		return errors.New("Build exited with non-zero status: " + strconv.FormatInt(status, 10))
	}
	fmt.Fprintln(out, "Build exited with status "+strconv.FormatInt(status, 10))

	// @TODO allow configuration
	dockerComposeRelFilePath := "docker-compose.yml"
	dockerComposeFilePath := os.Getenv("HOME") + "/project/" + dockerComposeRelFilePath

	// Set up docker-compose up
	fmt.Fprintln(out, "Preparing to start project...")
	resp, err = cli.ContainerCreate(
		ctx, &container.Config{
			Image:      DockerComposeVersion,
			WorkingDir: "/build/project",
			Env:        []string{"HOME=/build"},
			Cmd: []string{
				// set project name
				"-p", d.project,
				"up",
			},
		},
		&container.HostConfig{
			AutoRemove: true,
			Binds: []string{
				dockerComposeFilePath + ":/build/project/docker-compose.yml",
				"/var/run/docker.sock:/var/run/docker.sock",
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

	fmt.Fprintln(out, "Starting up project...")
	return cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
}

// dockerBuild builds project from Dockerfile and deploys it
func dockerBuild(d *Deployment, cli *docker.Client, out io.Writer) error {
	fmt.Fprintln(out, "Building Dockerfile project...")
	ctx := context.Background()
	buildCtx := bytes.NewBuffer(nil)
	err := common.BuildTar(Directory, buildCtx)
	if err != nil {
		return err
	}

	// @TODO: support configuration
	dockerFilePath := "Dockerfile"

	// Build image
	imageName := d.project + "-image"
	buildResp, err := cli.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		Tags:           []string{imageName},
		Remove:         true,
		Dockerfile:     dockerFilePath,
		SuppressOutput: false,
	})
	if err != nil {
		return err
	}

	// Output build progress
	stop := make(chan struct{})
	common.FlushRoutine(out, buildResp.Body, stop)
	close(stop)
	buildResp.Body.Close()

	fmt.Fprintf(out, "%s (%s) build has exited\n", imageName, buildResp.OSType)

	// Create container from image
	containerResp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image: imageName,
		},
		&container.HostConfig{
			AutoRemove: true,
		}, nil, d.project,
	)
	if err != nil {
		if strings.Contains(err.Error(), "No such image") {
			return errors.New("Image build was unsuccessful")
		}
		return err
	}
	if len(containerResp.Warnings) > 0 {
		warnings := strings.Join(containerResp.Warnings, "\n")
		return errors.New(warnings)
	}

	fmt.Fprintln(out, "Starting up project in container "+d.project+"...")
	return cli.ContainerStart(ctx, containerResp.ID, types.ContainerStartOptions{})
}

// herokuishBuild uses the Herokuish tool to use Heroku's official buidpacks
// to build the user project.
func herokuishBuild(d *Deployment, cli *docker.Client, out io.Writer) error {
	fmt.Fprintln(out, "Setting up herokuish...")
	ctx := context.Background()

	// Configure herokuish container to build project when run
	resp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image: HerokuishVersion,
			Cmd:   []string{"/build"},
		},
		&container.HostConfig{
			Binds: []string{
				// "/tmp/app" is the directory herokuish looks
				// for during a build, so mount project there
				os.Getenv("HOME") + "/project:/tmp/app",
			},
		}, nil, BuildStageName,
	)
	if err != nil {
		return err
	}
	if len(resp.Warnings) > 0 {
		fmt.Fprintln(out, "Warnings encountered on herokuish setup.")
		warnings := strings.Join(resp.Warnings, "\n")
		return errors.New(warnings)
	}

	// Start the herokuish container to build project
	fmt.Fprintln(out, "Building project...")
	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// Attach logs and report build progress until container exits
	reader, err := d.Logs(cli, LogOptions{Container: resp.ID, Stream: true})
	if err != nil {
		return err
	}
	stop := make(chan struct{})
	go common.FlushRoutine(out, reader, stop)
	status, err := cli.ContainerWait(ctx, resp.ID)
	close(stop)
	reader.Close()
	if err != nil {
		return err
	}
	if status != 0 {
		return errors.New("Build exited with non-zero status: " + strconv.FormatInt(status, 10))
	}
	fmt.Fprintln(out, "Build exited with status "+strconv.FormatInt(status, 10))

	// Save build as new image and create a container
	fmt.Fprintln(out, "Saving build...")
	_, err = cli.ContainerCommit(ctx, resp.ID, types.ContainerCommitOptions{
		Reference: "inertia-build",
	})
	if err != nil {
		return err
	}
	resp, err = cli.ContainerCreate(ctx, &container.Config{
		Image: "inertia-build:latest",
		// Currently, only start the standard "web" process
		Cmd: []string{"/start", "web"},
	}, nil, nil, d.project)
	if err != nil {
		return err
	}
	if len(resp.Warnings) > 0 {
		fmt.Fprintln(out, "Warnings encountered on herokuish startup.")
		warnings := strings.Join(resp.Warnings, "\n")
		return errors.New(warnings)
	}

	fmt.Fprintln(out, "Starting up project in container "+d.project+"...")
	return cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
}
