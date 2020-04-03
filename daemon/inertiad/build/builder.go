package build

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/cfg"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/log"
)

// ContainerBuilder builds projects and returns a callback that can be used to deploy the project.
// No relation to Bob the Builder, though a Bob did write this.
type ContainerBuilder interface {
	Build(string, Config, *docker.Client, io.Writer) (func() error, error)
	GetBuildStageName() string
	StopContainers(*docker.Client, io.Writer) error
	Prune(*docker.Client, io.Writer) error
	PruneAll(*docker.Client, io.Writer) error
}

// ProjectBuilder builds projects and returns a callback that can be used to deploy the project.
// No relation to Bob the Builder, though a Bob did write this.
type ProjectBuilder func(Config, *docker.Client, io.Writer) (func() error, error)

// Builder manages build tools and executes builds
type Builder struct {
	buildStageName       string
	dockerComposeVersion string
	stopper              containers.ContainerStopper

	builders map[string]ProjectBuilder
}

// NewBuilder creates a builder with given configuration
func NewBuilder(conf cfg.Config, stopper containers.ContainerStopper) *Builder {
	b := &Builder{
		buildStageName:       "build",
		dockerComposeVersion: conf.DockerComposeVersion,
		stopper:              stopper,
	}
	b.builders = map[string]ProjectBuilder{
		"dockerfile":     b.dockerBuild,
		"docker-compose": b.dockerCompose,
	}
	return b
}

// GetBuildStageName returns the name of the intermediary container used to
// build projects
func (b *Builder) GetBuildStageName() string { return b.buildStageName }

// StopContainers stops containers and cleans up assets
func (b *Builder) StopContainers(docker *docker.Client, out io.Writer) error {
	return b.stopper(docker, out)
}

// Prune cleans up Dokcer assets
func (b *Builder) Prune(docker *docker.Client, out io.Writer) error {
	return containers.Prune(docker)
}

// PruneAll forcibly removes Docker assets
func (b *Builder) PruneAll(docker *docker.Client, out io.Writer) error {
	return containers.PruneAll(docker, b.dockerComposeVersion)
}

// Config contains parameters required for builds to execute
type Config struct {
	Name string

	BuildFilePath    string
	BuildDirectory   string
	PersistDirectory string

	EnvValues []string
}

// Build executes build and deploy
func (b *Builder) Build(buildType string, d Config,
	cli *docker.Client, out io.Writer) (func() error, error) {
	// Use the appropriate build method
	builder, found := b.builders[strings.ToLower(buildType)]
	if !found {
		// @todo: attempt a guess at project type instead
		fmt.Println(out, "Unknown project type "+buildType)
		fmt.Println(out, "Defaulting to docker-compose build")
		builder = b.dockerCompose
	}

	// Build project
	reportDeployInit(buildType, d.Name, out)
	deploy, err := builder(d, cli, out)
	if err != nil {
		return func() error { return nil }, err
	}

	// Return the deploy callback
	return deploy, nil
}

// dockerCompose builds and runs project using docker-compose -
// the following code performs the bash equivalent of:
//
//    docker run -d \
// 	    -v /var/run/docker.sock:/var/run/docker.sock \
// 	    -v $HOME:/build \
// 	    -w="/build/project" \
// 	    docker/compose:latest up --build
//
// This starts a new container running a docker-compose image for
// the sole purpose of building the project. This container is
// separate from the daemon and the user's project, and is the
// second container to require access to the docker socket.
// See https://cloud.google.com/community/tutorials/docker-compose-on-container-optimized-os
func (b *Builder) dockerCompose(d Config, cli *docker.Client,
	out io.Writer) (func() error, error) {
	fmt.Fprintln(out, "Setting up docker-compose...")
	ctx := context.Background()

	dockercomposeFilePath := "docker-compose.yml"
	if d.BuildFilePath != "" {
		dockercomposeFilePath = d.BuildFilePath
	}

	// set up bindings
	binds := []string{
		getTrueDirectory(d.BuildDirectory) + ":/build",
		"/var/run/docker.sock:/var/run/docker.sock",
	}
	if d.PersistDirectory != "" {
		binds = append(binds, getTrueDirectory(d.PersistDirectory)+":/persist")
	}

	// set up docker-compose runner
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
			Binds:      binds,
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
	reportProjectBuildBegin(d.Name, out)
	if err := containers.StartAndWait(cli, resp.ID, out); err != nil {
		return nil, err
	}
	reportProjectBuildComplete(d.Name, out)

	// @TODO allow configuration
	var (
		dockerComposeRelFilePath = "docker-compose.yml"
		dockerComposeFilePath    = path.Join(
			getTrueDirectory(d.BuildDirectory), dockerComposeRelFilePath,
		)
	)

	// Set up docker-compose up
	reportProjectContainerCreateBegin(d.Name, out)
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
	reportProjectContainerCreateComplete(d.Name, out)

	return func() error { return b.run(ctx, cli, d.Name, resp.ID, out) }, nil
}

// dockerBuild builds project from Dockerfile, and returns a callback function to deploy it
func (b *Builder) dockerBuild(d Config, cli *docker.Client,
	out io.Writer) (func() error, error) {
	var (
		ctx      = context.Background()
		buildCtx = bytes.NewBuffer(nil)
	)

	// Create build context
	if err := buildTar(d.BuildDirectory, buildCtx); err != nil {
		return nil, err
	}

	// @TODO: support configuration
	dockerFilePath := "Dockerfile"
	if d.BuildFilePath != "" {
		dockerFilePath = d.BuildFilePath
	}

	// Build image
	reportProjectBuildBegin(d.Name, out)
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
	stop := make(chan struct{})
	log.FlushRoutine(out, buildResp.Body, stop)
	close(stop)
	buildResp.Body.Close()
	// Get image details - this will check if image build was successful
	image, _, err := cli.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return nil, fmt.Errorf("image build failed: %s", err.Error())
	}
	portMap := nat.PortMap{}
	for p := range image.Config.ExposedPorts {
		portMap[p] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: p.Port()}}
	}
	reportProjectBuildComplete(d.Name, out)

	// set up bindings
	binds := []string{}
	if d.PersistDirectory != "" {
		binds = append(binds, getTrueDirectory(d.PersistDirectory)+":/persist")
	}

	// Create container from image
	reportProjectContainerCreateBegin(d.Name, out)
	containerResp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image: imageName,
			Env:   d.EnvValues,
		},
		&container.HostConfig{
			Binds:        binds,
			PortBindings: portMap,
		}, nil, d.Name)
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
	reportProjectContainerCreateComplete(d.Name, out)

	return func() error { return b.run(ctx, cli, d.Name, containerResp.ID, out) }, nil
}

// run starts project and tracks all active project containers
func (b *Builder) run(ctx context.Context, client *docker.Client, name, id string, out io.Writer) error {
	reportProjectStartup(name, out)
	return client.ContainerStart(ctx, id, types.ContainerStartOptions{})
}
