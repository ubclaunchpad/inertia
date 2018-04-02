package project

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertia/auth"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	// DockerComposeVersion is the version of docker-compose used
	DockerComposeVersion = "docker/compose:1.19.0"

	// HerokuishVersion is the version of Herokuish used
	HerokuishVersion = "gliderlabs/herokuish:v0.4.0"

	// Directory specifies the location of deployed project
	Directory = "/app/host/project"
)

// Deployer does great deploys
type Deployer interface {
	Deploy(DeployOptions, *docker.Client, io.Writer) error
	Down(*docker.Client, io.Writer) error
	Destroy(*docker.Client, io.Writer) error

	Logs(LogOptions, *docker.Client) (io.ReadCloser, error)
	GetStatus(*docker.Client) (*DeploymentStatus, error)

	SetConfig(DeploymentConfig)
	GetBranch() string
	CompareRemotes(string) error
}

// Deployment represents the deployed project
type Deployment struct {
	project   string
	branch    string
	buildType string

	repo *git.Repository
	auth ssh.AuthMethod
	mux  sync.Mutex
}

// DeploymentConfig is used to configure Deployment
type DeploymentConfig struct {
	ProjectName string
	BuildType   string
	RemoteURL   string
	Branch      string
	PemFilePath string
}

// NewDeployment creates a new deployment
func NewDeployment(cfg DeploymentConfig, out io.Writer) (*Deployment, error) {
	pemFile, err := os.Open(cfg.PemFilePath)
	if err != nil {
		return nil, err
	}
	authMethod, err := auth.GetGithubKey(pemFile)
	if err != nil {
		return nil, err
	}
	repo, err := initializeRepository(cfg.RemoteURL, cfg.Branch, authMethod, out)
	if err != nil {
		return nil, err
	}

	return &Deployment{
		project:   cfg.ProjectName,
		branch:    cfg.Branch,
		buildType: cfg.BuildType,
		auth:      authMethod,
		repo:      repo,
	}, nil
}

// SetConfig updates the deployment's configuration. Only supports
// ProjectName, Branch, and BuildType for now.
func (d *Deployment) SetConfig(cfg DeploymentConfig) {
	if cfg.ProjectName != "" {
		d.project = cfg.ProjectName
	}
	if cfg.Branch != "" {
		d.branch = cfg.Branch
	}
	if cfg.BuildType != "" {
		d.buildType = cfg.BuildType
	}
}

// DeployOptions is used to configure how the deployment handles the deploy
type DeployOptions struct {
	SkipUpdate bool
}

// Deploy will update, build, and deploy the project
func (d *Deployment) Deploy(opts DeployOptions, cli *docker.Client, out io.Writer) error {
	d.mux.Lock()
	defer d.mux.Unlock()
	fmt.Println(out, "Preparing to deploy project")

	// Update repository
	if !opts.SkipUpdate {
		err := updateRepository(Directory, d.repo, d.branch, d.auth, out)
		if err != nil {
			return err
		}
	}

	// Kill active project containers if there are any
	err := stopActiveContainers(cli, out)
	if err != nil {
		return err
	}

	// Use the appropriate build method
	switch d.buildType {
	case "herokuish":
		return d.herokuishBuild(cli, out)
	default:
		return d.dockerCompose(cli, out)
	}
}

// Down shuts down the deployment
func (d *Deployment) Down(cli *docker.Client, out io.Writer) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	// Error if no project containers are active, but try to kill
	// everything anyway in case the docker-compose image is still
	// active
	_, err := getActiveContainers(cli)
	if err != nil {
		killErr := stopActiveContainers(cli, out)
		if killErr != nil {
			println(err)
		}
		return err
	}
	return stopActiveContainers(cli, out)
}

// Destroy shuts down the deployment and removes the repository
func (d *Deployment) Destroy(cli *docker.Client, out io.Writer) error {
	d.Down(cli, out)

	d.mux.Lock()
	defer d.mux.Unlock()
	return common.RemoveContents(Directory)
}

// LogOptions is used to configure retrieved container logs
type LogOptions struct {
	Container string
	Stream    bool
}

// Logs get logs ;)
func (d *Deployment) Logs(opts LogOptions, cli *docker.Client) (io.ReadCloser, error) {
	ctx := context.Background()
	return cli.ContainerLogs(ctx, opts.Container, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     opts.Stream,
		Timestamps: true,
	})
}

// DeploymentStatus lists details about the deployed project
type DeploymentStatus struct {
	Branch               string
	CommitHash           string
	CommitMessage        string
	Containers           []string
	BuildContainerActive bool
}

// GetStatus returns the status of the deployment
func (d *Deployment) GetStatus(cli *docker.Client) (*DeploymentStatus, error) {
	// Get repository status
	head, err := d.repo.Head()
	if err != nil {
		return nil, err
	}
	commit, err := d.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	// Get containers, filtering out non-project containers
	buildContainerActive := false
	containers, err := getActiveContainers(cli)
	if err != nil && err != ErrNoContainers {
		return nil, err
	}
	ignore := map[string]bool{
		"/inertia-daemon": true,
		"/build":          true,
	}
	activeContainers := make([]string, 0)
	for _, container := range containers {
		if !ignore[container.Names[0]] {
			activeContainers = append(activeContainers, container.Names[0])
		} else {
			if container.Names[0] == "/docker-compose" {
				buildContainerActive = true
			}
		}
	}

	return &DeploymentStatus{
		Branch:               strings.TrimSpace(head.Name().Short()),
		CommitHash:           strings.TrimSpace(head.Hash().String()),
		CommitMessage:        strings.TrimSpace(commit.Message),
		Containers:           activeContainers,
		BuildContainerActive: buildContainerActive,
	}, nil
}

// GetBranch returns the currently deployed branch
func (d *Deployment) GetBranch() string {
	return d.branch
}

// CompareRemotes will compare the remote of the deployment
// with given remote URL and return nil if they match
func (d *Deployment) CompareRemotes(remoteURL string) error {
	remotes, err := d.repo.Remotes()
	if err != nil {
		return err
	}
	localRemoteURL := common.GetSSHRemoteURL(remotes[0].Config().URLs[0])
	if localRemoteURL != common.GetSSHRemoteURL(remoteURL) {
		return errors.New("The given remote URL does not match that of the repository in\nyour remote - try 'inertia [REMOTE] reset'")
	}
	return nil
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
func (d *Deployment) dockerCompose(cli *docker.Client, out io.Writer) error {
	fmt.Fprintln(out, "Setting up docker-compose...")
	ctx := context.Background()
	resp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image:      DockerComposeVersion,
			WorkingDir: "/build/project",
			Env:        []string{"HOME=/build"},
			Cmd: []string{
				"-p", d.project,
				"up",
				"--build",
			},
		},
		&container.HostConfig{
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
				os.Getenv("HOME") + ":/build",
			},
		}, nil, "build",
	)
	if err != nil {
		return err
	}
	if len(resp.Warnings) > 0 {
		warnings := strings.Join(resp.Warnings, "\n")
		return errors.New(warnings)
	}

	fmt.Fprintln(out, "Building and starting up project...")
	return cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
}

// herokuishBuild uses the Herokuish tool to use Heroku's official buidpacks
// to build the user project.
func (d *Deployment) herokuishBuild(cli *docker.Client, out io.Writer) error {
	fmt.Fprintln(out, "Setting up herokuish...")
	ctx := context.Background()
	resp, err := cli.ContainerCreate(
		ctx, &container.Config{
			Image:        HerokuishVersion,
			AttachStdout: true,
			Cmd:          []string{"/build"},
		},
		&container.HostConfig{
			Binds: []string{
				Directory + ":/tmp/app",
			},
		}, nil, "build",
	)
	if err != nil {
		return err
	}
	if len(resp.Warnings) > 0 {
		fmt.Fprintln(out, "Warnings encountered on herokuish setup.")
		warnings := strings.Join(resp.Warnings, "\n")
		return errors.New(warnings)
	}

	fmt.Fprintln(out, "Building project...")
	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// Attach logs and report build progress until container exits
	reader, err := d.Logs(LogOptions{Container: resp.ID, Stream: true}, cli)
	if err != nil {
		return err
	}
	stop := make(chan bool)
	go common.FlushRoutine(out, reader, stop)
	status, err := cli.ContainerWait(ctx, resp.ID)
	stop <- true
	if err != nil {
		return err
	}
	if status != 0 {
		return errors.New("Build exited with non-zero status: " + strconv.FormatInt(status, 10))
	}

	// Save build and deploy image
	fmt.Fprintln(out, "Saving build...")
	_, err = cli.ContainerCommit(ctx, resp.ID, types.ContainerCommitOptions{
		Reference: "inertia-build",
		Config: &container.Config{
			AttachStdout: true,
			Cmd:          []string{"/start"},
		},
	})
	fmt.Fprintln(out, "Starting up project...")
	return cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
}
