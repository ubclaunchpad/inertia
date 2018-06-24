package build

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/cfg"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
)

// ProjectBuilder builds projects and returns a callback that can be used to deploy the project.
// No relation to Bob the Builder, though a Bob did write this.
type ProjectBuilder func(*Config, *docker.Client, io.Writer) (func() error, error)

// Builder manages build tools and executes builds
type Builder struct {
	buildStageName       string
	dockerComposeVersion string
	herokuishVersion     string

	builders map[string]ProjectBuilder
}

// NewBuilder creates a builder with given configuration
func NewBuilder(conf cfg.Config) *Builder {
	b := &Builder{
		buildStageName:       "build",
		dockerComposeVersion: conf.DockerComposeVersion,
		herokuishVersion:     conf.HerokuishVersion,
	}
	b.builders = map[string]ProjectBuilder{
		"herokuish":      b.herokuishBuild,
		"dockerfile":     b.dockerBuild,
		"docker-compose": b.dockerCompose,
	}
	return b
}

// GetBuildStageName returns the name of the intermediary container used to
// build projects
func (b *Builder) GetBuildStageName() string { return b.buildStageName }

// Config contains parameters required for builds to execute
type Config struct {
	Name string
	Type string

	BuildFilePath  string
	BuildDirectory string

	EnvValues []string
}

// Build executes build and deploy
func (b *Builder) Build(buildType string, d *Config,
	cli *docker.Client, out io.Writer) (func() error, error) {
	// Use the appropriate build method
	builder, found := b.builders[strings.ToLower(d.Type)]
	if !found {
		// @todo: attempt a guess at project type instead
		fmt.Println(out, "Unknown project type "+d.Type)
		fmt.Println(out, "Defaulting to docker-compose build")
		builder = b.dockerCompose
	}

	// Deploy project
	deploy, err := builder(d, cli, out)
	if err != nil {
		return func() error { return nil }, err
	}
	return deploy, nil
}

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
func (b *Builder) dockerCompose(d *Config, cli *docker.Client,
	out io.Writer) (func() error, error) {
	fmt.Fprintln(out, "Setting up docker-compose...")
	ctx := context.Background()

	dockercomposeFilePath := "docker-compose.yml"
	if d.BuildFilePath != "" {
		dockercomposeFilePath = d.BuildFilePath
	}

	resp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image:      b.dockerComposeVersion,
			WorkingDir: "/build",
			Cmd: []string{
				"-p", d.Name,
				"-f", dockercomposeFilePath,
				"build",
			},
			Env: d.EnvValues,
		},
		&container.HostConfig{
			AutoRemove: true,
			Binds: []string{
				getTrueDirectory(d.BuildDirectory) + ":/build",
				"/var/run/docker.sock:/var/run/docker.sock",
			},
		}, nil, b.buildStageName,
	)
	if err != nil {
		return nil, err
	}
	if len(resp.Warnings) > 0 {
		fmt.Fprintln(out, "Warnings encountered on docker-compose build.")
		warnings := strings.Join(resp.Warnings, "\n")
		return nil, errors.New(warnings)
	}

	// Start container to build project
	fmt.Fprintln(out, "Building project...")
	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}
	stop := make(chan struct{})
	go containers.StreamContainerLogs(cli, resp.ID, out, stop)
	status, err := cli.ContainerWait(ctx, resp.ID)
	close(stop)
	if err != nil {
		return nil, err
	}
	if status != 0 {
		return nil, errors.New("Build exited with non-zero status: " + strconv.FormatInt(status, 10))
	}
	fmt.Fprintln(out, "Build exited with status "+strconv.FormatInt(status, 10))

	// @TODO allow configuration
	dockerComposeRelFilePath := "docker-compose.yml"
	dockerComposeFilePath := path.Join(
		getTrueDirectory(d.BuildDirectory), dockerComposeRelFilePath,
	)

	// Set up docker-compose up
	fmt.Fprintln(out, "Preparing to start project...")
	resp, err = cli.ContainerCreate(
		ctx, &container.Config{
			Image:      b.dockerComposeVersion,
			WorkingDir: "/build",
			Cmd: []string{
				"-p", d.Name,
				"-f", dockercomposeFilePath,
				"up",
			},
			Env: d.EnvValues,
		},
		&container.HostConfig{
			AutoRemove: true,
			Binds: []string{
				dockerComposeFilePath + ":/build/docker-compose.yml",
				"/var/run/docker.sock:/var/run/docker.sock",
			},
		}, nil, "docker-compose",
	)
	if err != nil {
		return nil, err
	}
	if len(resp.Warnings) > 0 {
		warnings := strings.Join(resp.Warnings, "\n")
		return nil, errors.New(warnings)
	}

	return func() error {
		fmt.Fprintln(out, "Starting up project...")
		return cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	}, nil
}

// dockerBuild builds project from Dockerfile, and returns a callback function to deploy it
func (b *Builder) dockerBuild(d *Config, cli *docker.Client,
	out io.Writer) (func() error, error) {
	fmt.Fprintln(out, "Building Dockerfile project...")
	ctx := context.Background()
	buildCtx := bytes.NewBuffer(nil)
	err := buildTar(d.BuildDirectory, buildCtx)
	if err != nil {
		return nil, err
	}

	// @TODO: support configuration
	dockerFilePath := "Dockerfile"
	if d.BuildFilePath != "" {
		dockerFilePath = d.BuildFilePath
	}

	// Build image
	imageName := "inertia-build/" + d.Name
	buildResp, err := cli.ImageBuild(
		ctx, buildCtx, types.ImageBuildOptions{
			Tags:           []string{imageName},
			Remove:         true,
			Dockerfile:     dockerFilePath,
			SuppressOutput: false,
		},
	)
	if err != nil {
		return nil, err
	}

	// Output build progress
	stop := make(chan struct{})
	log.FlushRoutine(out, buildResp.Body, stop)
	close(stop)
	buildResp.Body.Close()
	fmt.Fprintf(out, "%s (%s) build has exited\n", imageName, buildResp.OSType)

	// Create container from image
	containerResp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image: imageName,
			Env:   d.EnvValues,
		},
		&container.HostConfig{
			AutoRemove: true,
		}, nil, d.Name,
	)
	if err != nil {
		if strings.Contains(err.Error(), "No such image") {
			return nil, errors.New("Image build was unsuccessful")
		}
		return nil, err
	}
	if len(containerResp.Warnings) > 0 {
		warnings := strings.Join(containerResp.Warnings, "\n")
		return nil, errors.New(warnings)
	}

	return func() error {
		fmt.Fprintln(out, "Starting up project in container "+d.Name+"...")
		return cli.ContainerStart(ctx, containerResp.ID, types.ContainerStartOptions{})
	}, nil
}

// herokuishBuild uses the Herokuish tool to use Heroku's official buidpacks
// to build the user project.
func (b *Builder) herokuishBuild(d *Config, cli *docker.Client,
	out io.Writer) (func() error, error) {
	fmt.Fprintln(out, "Setting up herokuish...")
	ctx := context.Background()

	// Configure herokuish container to build project when run
	resp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image: b.herokuishVersion,
			Cmd:   []string{"/build"},
			Env:   d.EnvValues,
		},
		&container.HostConfig{
			Binds: []string{
				// "/tmp/app" is the directory herokuish looks
				// for during a build, so mount project there
				getTrueDirectory(d.BuildDirectory) + ":/tmp/app",
			},
		}, nil, b.buildStageName,
	)
	if err != nil {
		return nil, err
	}
	if len(resp.Warnings) > 0 {
		fmt.Fprintln(out, "Warnings encountered on herokuish setup.")
		warnings := strings.Join(resp.Warnings, "\n")
		return nil, errors.New(warnings)
	}

	// Start the herokuish container to build project
	fmt.Fprintln(out, "Building project...")
	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	// Attach logs and report build progress until container exits
	stop := make(chan struct{})
	go containers.StreamContainerLogs(cli, resp.ID, out, stop)
	status, err := cli.ContainerWait(ctx, resp.ID)
	close(stop)
	if err != nil {
		return nil, err
	}
	if status != 0 {
		return nil, errors.New("Build exited with non-zero status: " + strconv.FormatInt(status, 10))
	}
	fmt.Fprintln(out, "Build exited with status "+strconv.FormatInt(status, 10))

	// Save build as new image and create a container
	imgName := "inertia-build/" + d.Name
	fmt.Fprintln(out, "Saving build...")
	_, err = cli.ContainerCommit(ctx, resp.ID, types.ContainerCommitOptions{
		Reference: imgName,
	})
	if err != nil {
		return nil, err
	}
	resp, err = cli.ContainerCreate(ctx, &container.Config{
		Image: imgName + ":latest",
		// Currently, only start the standard "web" process
		// @todo more processes
		Cmd: []string{"/start", "web"},
		Env: d.EnvValues,
	}, nil, nil, d.Name)
	if err != nil {
		return nil, err
	}
	if len(resp.Warnings) > 0 {
		fmt.Fprintln(out, "Warnings encountered on herokuish startup.")
		warnings := strings.Join(resp.Warnings, "\n")
		return nil, errors.New(warnings)
	}

	fmt.Fprintln(out, "Starting up project in container "+d.Name+"...")
	return func() error {
		return cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	}, nil
}
