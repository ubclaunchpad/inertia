package project

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
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

// Deployment represents the deployed project
type Deployment struct {
	Project string
	Branch  string
	Type    string

	repo *git.Repository
	auth ssh.AuthMethod
	mux  sync.Mutex
}

// DeploymentStatus lists details about the deployed project
type DeploymentStatus struct {
	Branch               string
	CommitHash           string
	CommitMessage        string
	Containers           []string
	BuildContainerActive bool
}

// NewDeployment creates a new deployment
func NewDeployment(projectName, remoteURL, branch string, out io.Writer) (*Deployment, error) {
	pemFile, err := os.Open(auth.DaemonGithubKeyLocation)
	if err != nil {
		return nil, err
	}
	authMethod, err := auth.GetGithubKey(pemFile)
	if err != nil {
		return nil, err
	}
	repo, err := initializeRepository(remoteURL, branch, authMethod, out)
	if err != nil {
		return nil, err
	}

	return &Deployment{
		Project: projectName,
		Branch:  branch,
		auth:    authMethod,
		repo:    repo,
	}, nil
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
		killErr := killActiveContainers(cli, out)
		if killErr != nil {
			println(err)
		}
		return err
	}
	return killActiveContainers(cli, out)
}

// Destroy shuts down the deployment and removes the repository
func (d *Deployment) Destroy(cli *docker.Client) error {
	d.Down(cli, nil)

	d.mux.Lock()
	defer d.mux.Unlock()
	return common.RemoveContents(Directory)
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

// Logs get logs ;)
func (d *Deployment) Logs(container string, follow bool, cli *docker.Client) (io.ReadCloser, error) {
	ctx := context.Background()
	return cli.ContainerLogs(ctx, container, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: true,
	})
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
		"/docker-compose": true,
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
		Branch:               head.Name().Short(),
		CommitHash:           head.Hash().String(),
		CommitMessage:        commit.Message,
		Containers:           activeContainers,
		BuildContainerActive: buildContainerActive,
	}, nil
}

// Deploy will update, build, and deploy the project
func (d *Deployment) Deploy(cli *docker.Client, out io.Writer) error {
	d.mux.Lock()
	defer d.mux.Unlock()
	fmt.Println(out, "Deploying repository...")

	// Update repository
	err := updateRepository(Directory, d.repo, d.Branch, d.auth, out)
	if err != nil {
		return err
	}

	// Kill active project containers if there are any
	err = killActiveContainers(cli, out)
	if err != nil {
		return err
	}

	return d.dockerCompose(cli, out)
}

func (d *Deployment) dockerCompose(cli *docker.Client, out io.Writer) error {
	// Build and run project using docker-compose - the following code
	// performs the bash equivalent of:
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
				"-p", d.Project,
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
}
