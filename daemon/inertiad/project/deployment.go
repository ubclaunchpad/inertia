package project

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/build"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/git"
	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// Builder builds projects and returns a callback that can be used to deploy the project.
// No relation to Bob the Builder, though a Bob did write this.
type Builder interface {
	Build(string, *build.Config, *docker.Client, io.Writer) (func() error, error)
	GetBuildStageName() string
}

// Deployer manages the deployed user project
type Deployer interface {
	Deploy(*docker.Client, io.Writer, DeployOptions) error
	Down(*docker.Client, io.Writer) error
	Destroy(*docker.Client, io.Writer) error

	GetStatus(*docker.Client) (*common.DeploymentStatus, error)
	SetConfig(DeploymentConfig)
	GetBranch() string
	CompareRemotes(string) error

	GetDataManager() (*DeploymentDataManager, bool)
}

// Deployment represents the deployed project
type Deployment struct {
	directory string

	project       string
	branch        string
	buildType     string
	buildFilePath string

	builder          Builder
	containerStopper containers.ContainerStopper

	repo *gogit.Repository
	auth ssh.AuthMethod
	mux  sync.Mutex

	dataManager *DeploymentDataManager
}

// DeploymentConfig is used to configure Deployment
type DeploymentConfig struct {
	ProjectDirectory string
	ProjectName      string
	BuildType        string
	BuildFilePath    string
	RemoteURL        string
	Branch           string
	PemFilePath      string
	DatabasePath     string
}

// NewDeployment creates a new deployment
func NewDeployment(builder Builder, cfg DeploymentConfig, out io.Writer) (*Deployment, error) {
	common.RemoveContents(cfg.ProjectDirectory)

	// Set up git repository
	pemFile, err := os.Open(cfg.PemFilePath)
	if err != nil {
		return nil, err
	}
	authMethod, err := crypto.GetGithubKey(pemFile)
	if err != nil {
		return nil, err
	}
	repo, err := git.InitializeRepository(
		cfg.ProjectDirectory, cfg.RemoteURL, cfg.Branch, authMethod, out,
	)
	if err != nil {
		return nil, err
	}

	// Set up deployment database
	manager, err := newDataManager(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}

	// Create deployment
	return &Deployment{
		// Properties
		directory: cfg.ProjectDirectory,
		project:   cfg.ProjectName,
		branch:    cfg.Branch,
		buildType: cfg.BuildType,

		// Functions
		builder:          builder,
		containerStopper: containers.StopActiveContainers,

		// Repository
		auth: authMethod,
		repo: repo,

		// Persistent data manager
		dataManager: manager,
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
	if cfg.BuildFilePath != "" {
		d.buildFilePath = cfg.BuildFilePath
	}
}

// DeployOptions is used to configure how the deployment handles the deploy
type DeployOptions struct {
	SkipUpdate bool
}

// Deploy will update, build, and deploy the project
func (d *Deployment) Deploy(cli *docker.Client, out io.Writer,
	opts DeployOptions) error {
	d.mux.Lock()
	defer d.mux.Unlock()
	fmt.Println(out, "Preparing to deploy project")

	// Update repository
	if !opts.SkipUpdate {
		err := git.UpdateRepository(d.directory, d.repo, d.branch, d.auth, out)
		if err != nil {
			return err
		}
	}

	// Kill active project containers if there are any
	err := d.containerStopper(cli, out)
	if err != nil {
		return err
	}

	// Get config
	conf, err := d.GetBuildConfiguration()
	if err != nil {
		fmt.Println(err.Error())
	}

	// Build project
	deploy, err := d.builder.Build(strings.ToLower(d.buildType), conf, cli, out)
	if err != nil {
		return err
	}

	// Deploy
	return deploy()
}

// Down shuts down the deployment
func (d *Deployment) Down(cli *docker.Client, out io.Writer) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	// Error if no project containers are active, but try to kill
	// everything anyway in case the docker-compose image is still
	// active
	_, err := containers.GetActiveContainers(cli)
	if err != nil {
		killErr := d.containerStopper(cli, out)
		if killErr != nil {
			println(err)
		}
		return err
	}
	return d.containerStopper(cli, out)
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
func (d *Deployment) GetStatus(cli *docker.Client) (*common.DeploymentStatus, error) {
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
	c, err := containers.GetActiveContainers(cli)
	if err != nil && err != containers.ErrNoContainers {
		return nil, err
	}
	ignore := map[string]bool{
		"/inertia-daemon":                   true,
		"/" + d.builder.GetBuildStageName(): true,
	}
	activeContainers := make([]string, 0)
	for _, container := range c {
		if !ignore[container.Names[0]] {
			activeContainers = append(activeContainers, container.Names[0])
		} else {
			if container.Names[0] == "/docker-compose" {
				buildContainerActive = true
			}
		}
	}

	return &common.DeploymentStatus{
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
		Name:           d.project,
		Type:           d.buildType,
		BuildFilePath:  d.buildFilePath,
		BuildDirectory: d.directory,
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
