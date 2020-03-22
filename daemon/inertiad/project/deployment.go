package project

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/build"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/git"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/notify"
)

// Deployer manages the deployed user project
type Deployer interface {
	Deploy(*docker.Client, io.Writer, DeployOptions) (func() error, error)
	Initialize(cfg DeploymentConfig, out io.Writer) error
	Down(*docker.Client, io.Writer) error
	Destroy(*docker.Client, io.Writer) error
	Prune(*docker.Client, io.Writer) error
	GetStatus(*docker.Client) (api.DeploymentStatus, error)

	SetConfig(DeploymentConfig)
	GetBranch() string
	CompareRemotes(string) error

	UpdateContainerHistory(cli *docker.Client) error

	GetDataManager() (*DeploymentDataManager, bool)

	Watch(*docker.Client) (<-chan string, <-chan error)
}

// Deployment represents the deployed project
type Deployment struct {
	active           bool
	directory        string
	persistDirectory string

	project                string
	branch                 string
	buildType              string
	buildFilePath          string
	intermediaryContainers []string

	builder build.ContainerBuilder

	repo *gogit.Repository
	auth ssh.AuthMethod
	mux  sync.Mutex

	dataManager *DeploymentDataManager

	notifiers notify.Notifiers
}

// DeploymentConfig is used to configure Deployment
type DeploymentConfig struct {
	ProjectName            string
	BuildType              string
	BuildFilePath          string
	RemoteURL              string
	Branch                 string
	PemFilePath            string
	IntermediaryContainers []string

	// TODO: maybe improve format for generic notifiers
	SlackNotificationURL string
}

// DeploymentMetadata is used to store metadata relevant
// to the most recent deployment
type DeploymentMetadata struct {
	Hash            string
	ContainerID     string
	ContainerStatus string
	StartedAt       string
}

// NewDeployment creates a new deployment
func NewDeployment(
	projectDirectory string,
	persistDirectory string,

	databasePath string,
	databaseKeyPath string,

	builder build.ContainerBuilder,
) (*Deployment, error) {

	// Set up deployment database
	manager, err := NewDataManager(databasePath, databaseKeyPath)
	if err != nil {
		return nil, err
	}

	// Create deployment
	return &Deployment{
		directory:        projectDirectory,
		persistDirectory: persistDirectory,
		builder:          builder,
		dataManager:      manager,
	}, nil
}

// Initialize sets up deployment repository
func (d *Deployment) Initialize(cfg DeploymentConfig, out io.Writer) error {
	if cfg.RemoteURL == "" {
		return errors.New("remote URL is required for first setup")
	}

	d.SetConfig(cfg)

	// Retrieve authentication
	pemFile, err := os.Open(cfg.PemFilePath)
	if err != nil {
		return err
	}
	d.auth, err = crypto.GetGithubKey(pemFile)
	if err != nil {
		return err
	}

	// Remove existing git repo if there is one
	os.RemoveAll(filepath.Join(d.directory, ".git"))

	// Remove existing /persist data
	if d.persistDirectory != "" {
		os.RemoveAll(d.persistDirectory)
	}

	// Initialize repository
	d.repo, err = git.InitializeRepository(cfg.RemoteURL, git.RepoOptions{
		Directory: d.directory,
		Branch:    cfg.Branch,
		Auth:      d.auth,
	}, out)
	return err
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
	if cfg.BuildFilePath != "" {
		d.buildFilePath = cfg.BuildFilePath
	}
	d.intermediaryContainers = cfg.IntermediaryContainers

	// register notifiers
	if len(d.notifiers) == 0 {
		d.notifiers = notify.Notifiers{}
	}
	if cfg.SlackNotificationURL != "" {
		nt := notify.NewSlackNotifier(cfg.SlackNotificationURL)
		if !d.notifiers.Exists(nt) {
			d.notifiers = append(d.notifiers, nt)
		}
	}
}

// DeployOptions is used to configure how the deployment handles the deploy
type DeployOptions struct {
	SkipUpdate bool
}

// Deploy will update, build, and deploy the project
func (d *Deployment) Deploy(
	cli *docker.Client,
	out io.Writer,
	opts DeployOptions,
) (func() error, error) {
	d.mux.Lock()
	defer d.mux.Unlock()
	fmt.Println(out, "Preparing to deploy project")

	// Update repository
	if !opts.SkipUpdate {
		if err := git.UpdateRepository(d.repo, git.RepoOptions{
			Directory: d.directory,
			Branch:    d.branch,
			Auth:      d.auth,
		}, out); err != nil {
			return func() error { return nil }, err
		}
	}

	// Clean up
	d.builder.Prune(cli, out)

	// Kill active project containers if there are any
	d.active = false
	err := d.builder.StopContainers(cli, out)
	if err != nil {
		return func() error { return nil }, err
	}

	// Get config
	conf, err := d.GetBuildConfiguration()
	if err != nil {
		fmt.Fprintln(out, err.Error())
		fmt.Fprintln(out, "Continuing...")
	}

	// Build project
	deploy, err := d.builder.Build(strings.ToLower(d.buildType), *conf, cli, out)
	if err != nil {
		if notifyErr := d.notifiers.Notify(fmt.Sprintf("Build error: %s", err), notify.Options{
			Color: notify.Red,
		}); notifyErr != nil {
			fmt.Fprintln(out, notifyErr.Error())
		}
		return func() error { return nil }, err
	}

	// Send build complete slack notification
	if notifyErr := d.notifiers.Notify("Build completed", notify.Options{
		Color: notify.Green,
	}); notifyErr != nil {
		fmt.Fprintln(out, notifyErr.Error())
	}

	// Deploy
	return func() error {
		d.active = true
		return deploy()
	}, nil
}

// Down shuts down the deployment
func (d *Deployment) Down(cli *docker.Client, out io.Writer) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	// Error if no project containers are active, but try to kill
	// everything anyway in case the docker-compose image is still
	// active
	d.active = false
	_, err := containers.GetActiveContainers(cli)
	if err != nil {
		killErr := d.builder.StopContainers(cli, out)
		if killErr != nil {
			println(err)
		}
		return err
	}
	err = d.builder.StopContainers(cli, out)
	if err != nil {
		return err
	}

	// Do a lite prune
	d.builder.Prune(cli, out)
	return nil
}

// Prune clears unused Docker assets
func (d *Deployment) Prune(cli *docker.Client, out io.Writer) error {
	return d.builder.PruneAll(cli, out)
}

// Destroy shuts down the deployment and removes the repository
func (d *Deployment) Destroy(cli *docker.Client, out io.Writer) error {
	d.Down(cli, out)

	d.mux.Lock()
	defer d.mux.Unlock()
	err := d.dataManager.destroy()
	if err != nil {
		fmt.Fprint(out, "unable to clear database records: "+err.Error())
	}
	return common.RemoveContents(d.directory)
}

// GetStatus returns the status of the deployment
func (d *Deployment) GetStatus(cli *docker.Client) (api.DeploymentStatus, error) {
	var (
		activeContainers     = make([]string, 0)
		buildContainerActive = false
		ignore               = map[string]bool{
			"/inertia-daemon":                   true,
			"/" + d.builder.GetBuildStageName(): true,
		}
	)

	// No repository set up
	if d.repo == nil {
		return api.DeploymentStatus{Containers: activeContainers}, nil
	}

	// Get repository status
	head, err := d.repo.Head()
	if err != nil {
		return api.DeploymentStatus{Containers: activeContainers}, err
	}
	commit, err := d.repo.CommitObject(head.Hash())
	if err != nil {
		return api.DeploymentStatus{Containers: activeContainers}, err
	}

	// Get containers, filtering out non-project containers
	c, err := containers.GetActiveContainers(cli)
	if err != nil && err != containers.ErrNoContainers {
		return api.DeploymentStatus{Containers: activeContainers}, err
	}
	for _, container := range c {
		if !ignore[container.Names[0]] {
			activeContainers = append(activeContainers, container.Names[0])
		} else {
			if container.Names[0] == "/docker-compose" {
				buildContainerActive = true
			}
		}
	}

	return api.DeploymentStatus{
		Branch:               strings.TrimSpace(head.Name().Short()),
		CommitHash:           strings.TrimSpace(head.Hash().String()),
		CommitMessage:        strings.TrimSpace(commit.Message),
		BuildType:            strings.TrimSpace(d.buildType),
		Containers:           activeContainers,
		BuildContainerActive: buildContainerActive,
	}, nil
}

// GetBranch returns the currently deployed branch
func (d *Deployment) GetBranch() string {
	return d.branch
}

// CompareRemotes will compare the remote of the deployment  with given remote
// URL and return nil if they don't conflict
func (d *Deployment) CompareRemotes(remoteURL string) error {
	// Ignore if no remote given
	if remoteURL == "" {
		return nil
	}
	remotes, err := d.repo.Remotes()
	if err != nil {
		return err
	}
	localRemoteURL := common.GetSSHRemoteURL(remotes[0].Config().URLs[0])
	if localRemoteURL != common.GetSSHRemoteURL(remoteURL) {
		return errors.New("The given remote URL does not match that of the repository in\nyour remote - try 'inertia [remote] reset'")
	}
	return nil
}

// UpdateContainerHistory will update container bucket with recent build's
// metadata
func (d *Deployment) UpdateContainerHistory(cli *docker.Client) error {

	// Get project hash
	head, err := d.repo.Head()
	if err != nil {
		return fmt.Errorf("failed fetching repo head when updating container history: %s", err.Error())
	}
	// Retrieve container for recently deployed project
	ctx := context.Background()
	var recentlyBuiltContainer types.Container
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return fmt.Errorf("failure fetching list of containers: %s", err.Error())
	}
	for _, container := range containers {
		if container.Names[0] == d.project {
			recentlyBuiltContainer = container
		}
	}

	// Get container metadata
	var containerID string
	if len(recentlyBuiltContainer.ID) > 0 {
		containerID = recentlyBuiltContainer.ID
	}
	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failure fetching container metadata: %s", err.Error())
	}
	containerState := containerJSON.ContainerJSONBase.State // similar to running "docker inspect {container}"
	var containerStatus, containerStartedAtTime string
	if containerState != nil {
		containerStatus = containerState.Status
		containerStartedAtTime = containerState.StartedAt
	}

	metadata := DeploymentMetadata{
		Hash:            head.Hash().String(),
		ContainerID:     containerID,
		ContainerStatus: containerStatus,
		StartedAt:       containerStartedAtTime}

	// Update db with newly built container metadata
	err = d.dataManager.AddProjectBuildData(d.project, metadata)
	if err != nil {
		return fmt.Errorf("failure adding build metadata: %s", err.Error())
	}

	return nil
}

// GetDataManager returns the class managing deployment data
func (d *Deployment) GetDataManager() (manager *DeploymentDataManager, found bool) {
	if d.dataManager == nil {
		return nil, false
	}
	return d.dataManager, true
}

// GetBuildConfiguration returns the build used to build this project. Returns
// config without env values if error.
func (d *Deployment) GetBuildConfiguration() (*build.Config, error) {
	conf := &build.Config{
		Name:             d.project,
		BuildFilePath:    d.buildFilePath,
		BuildDirectory:   d.directory,
		PersistDirectory: d.persistDirectory,
	}
	if d.dataManager != nil {
		env, err := d.dataManager.GetEnvVariables(true)
		if err != nil {
			return conf, err
		}
		conf.EnvValues = env
	} else {
		return conf, errors.New("no data manager")
	}
	return conf, nil
}

// Watch watches for container stops
func (d *Deployment) Watch(client *docker.Client) (<-chan string, <-chan error) {
	var (
		ctx    = context.Background()
		logsCh = make(chan string)
		errCh  = make(chan error)
	)

	// Listen on channels
	go func() {
		defer close(errCh)

		// Only listen for die events
		eventsCh, eventsErrCh := client.Events(ctx,
			types.EventsOptions{Filters: filters.NewArgs(
				filters.KeyValuePair{Key: "event", Value: "die"}),
			})

		for {
			select {
			case err := <-eventsErrCh:
				if err != nil {
					errCh <- err
					break
				}

			case status := <-eventsCh:
				var containerName string
				if status.Actor.Attributes != nil {
					containerName = status.Actor.Attributes["name"]
				}

				if containerName != "" {
					logsCh <- fmt.Sprintf("container %s (%s) has stopped", containerName, status.ID[:11])
				} else {
					logsCh <- fmt.Sprintf("container %s has stopped", status.ID[:11])
				}

				if d.active {
					// Check if we should ignore this container's death
					var ignore bool
					if len(d.intermediaryContainers) > 0 && containerName == "" {
						for _, c := range d.intermediaryContainers {
							if containerName == c {
								ignore = true
							}
						}
					}

					// Shut down all containers if one stops while project is active
					if !ignore {
						d.active = false
						logsCh <- "container stoppage was unexpected, project is active"
						err := containers.StopActiveContainers(client, os.Stdout)
						if err != nil {
							logsCh <- ("error shutting down other active containers: " + err.Error())
						}
					}
				}
			}
		}
	}()

	return logsCh, errCh
}
