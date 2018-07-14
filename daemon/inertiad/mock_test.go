package main

import (
	"io"

	docker "github.com/docker/docker/client"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/project"
)

// This file contains mock implementations of interfaces used by this
// package for testing purposes.

// FakeDeployment is an implementation of the project.Deployer interface.
// Make sure to assign functions to each field that gets called or a nil
// pointer will be thrown.
type FakeDeployment struct {
	CompareRemotesFunc func(in1 string) error
	DeployFunc         func(in1 project.DeployOptions, in2 *docker.Client, in3 io.Writer) error
	DestroyFunc        func(in1 *docker.Client, in2 io.Writer) error
	DownFunc           func(in1 *docker.Client, in2 io.Writer) error
	GetBranchFunc      func() string
	GetStatusFunc      func(in1 *docker.Client) (*common.DeploymentStatus, error)
	UpdateEnvFunc      func(cli *docker.Client) error
	GetDataManagerFunc func() *project.DeploymentDataManager
}

func (f *FakeDeployment) Deploy(c *docker.Client, w io.Writer, o project.DeployOptions) error {
	return f.DeployFunc(o, c, w)
}

func (f *FakeDeployment) Down(c *docker.Client, w io.Writer) error {
	return f.DownFunc(c, w)
}

func (f *FakeDeployment) Prune(c *docker.Client, w io.Writer) error { return nil }

func (f *FakeDeployment) Destroy(c *docker.Client, w io.Writer) error {
	return f.DestroyFunc(c, w)
}

func (f *FakeDeployment) GetStatus(c *docker.Client) (*common.DeploymentStatus, error) {
	return f.GetStatusFunc(c)
}

func (f *FakeDeployment) SetConfig(project.DeploymentConfig) {}

func (f *FakeDeployment) GetBranch() string {
	return f.GetBranchFunc()
}

func (f *FakeDeployment) CompareRemotes(s string) error {
	return f.CompareRemotesFunc(s)
}

func (f *FakeDeployment) GetDataManager() (*project.DeploymentDataManager, bool) {
	return nil, false
}

func (f *FakeDeployment) UpdateContainerEnvironmentValues(cli *docker.Client) error {
	return nil
}
