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
	"github.com/ubclaunchpad/inertia/daemon/inertiad/auth"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/containers"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/git"
	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	// directory is the directory the user's deployed project is cloned in
	directory = os.Getenv("INERTIA_PROJECT_DIR") // "/app/host/inertia/project"
)

// Deployer does great deploys
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

	builders         map[string]Builder
	containerStopper containers.ContainerStopper

	repo *gogit.Repository
	auth ssh.AuthMethod
	mux  sync.Mutex

	dataManager *DeploymentDataManager
}

// DeploymentConfig is used to configure Deployment
type DeploymentConfig struct {
	ProjectName   string
	BuildType     string
	BuildFilePath string
	RemoteURL     string
	Branch        string
	PemFilePath   string
	DatabasePath  string
}

// NewDeployment creates a new deployment
func NewDeployment(cfg DeploymentConfig, out io.Writer) (*Deployment, error) {
	common.RemoveContents(directory)

	// Set up git repository
	pemFile, err := os.Open(cfg.PemFilePath)
	if err != nil {
		return nil, err
	}
	authMethod, err := auth.GetGithubKey(pemFile)
	if err != nil {
		return nil, err
	}
	repo, err := git.InitializeRepository(directory, cfg.RemoteURL, cfg.Branch, authMethod, out)
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
		directory: directory,
		project:   cfg.ProjectName,
		branch:    cfg.Branch,
		buildType: cfg.BuildType,

		// Functions
		builders: map[string]Builder{
			"herokuish":      herokuishBuild,
			"dockerfile":     dockerBuild,
			"docker-compose": dockerCompose,
		},
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
func (d *Deployment) Deploy(cli *docker.Client, out io.Writer, opts DeployOptions) error {
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

	// Use the appropriate build method
	builder, found := d.builders[strings.ToLower(d.buildType)]
	if !found {
		// @todo: attempt a guess at project type instead
		fmt.Println(out, "Unknown project type "+d.buildType)
		fmt.Println(out, "Defaulting to docker-compose build")
		builder = dockerCompose
	}

	// Kill active project containers if there are any
	err := d.containerStopper(cli, out)
	if err != nil {
		return err
	}

	// Deploy project
	deploy, err := builder(d, cli, out)
	if err != nil {
		return err
	}

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
		"/inertia-daemon":    true,
		"/" + BuildStageName: true,
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
