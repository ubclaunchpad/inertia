// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	io "io"
	sync "sync"

	client "github.com/docker/docker/client"
	common "github.com/ubclaunchpad/inertia/common"
	project "github.com/ubclaunchpad/inertia/daemon/inertiad/project"
)

type FakeDeployer struct {
	CompareRemotesStub        func(string) error
	compareRemotesMutex       sync.RWMutex
	compareRemotesArgsForCall []struct {
		arg1 string
	}
	compareRemotesReturns struct {
		result1 error
	}
	compareRemotesReturnsOnCall map[int]struct {
		result1 error
	}
	DeployStub        func(*client.Client, io.Writer, project.DeployOptions) (func() error, error)
	deployMutex       sync.RWMutex
	deployArgsForCall []struct {
		arg1 *client.Client
		arg2 io.Writer
		arg3 project.DeployOptions
	}
	deployReturns struct {
		result1 func() error
		result2 error
	}
	deployReturnsOnCall map[int]struct {
		result1 func() error
		result2 error
	}
	DestroyStub        func(*client.Client, io.Writer) error
	destroyMutex       sync.RWMutex
	destroyArgsForCall []struct {
		arg1 *client.Client
		arg2 io.Writer
	}
	destroyReturns struct {
		result1 error
	}
	destroyReturnsOnCall map[int]struct {
		result1 error
	}
	DownStub        func(*client.Client, io.Writer) error
	downMutex       sync.RWMutex
	downArgsForCall []struct {
		arg1 *client.Client
		arg2 io.Writer
	}
	downReturns struct {
		result1 error
	}
	downReturnsOnCall map[int]struct {
		result1 error
	}
	GetBranchStub        func() string
	getBranchMutex       sync.RWMutex
	getBranchArgsForCall []struct {
	}
	getBranchReturns struct {
		result1 string
	}
	getBranchReturnsOnCall map[int]struct {
		result1 string
	}
	GetDataManagerStub        func() (*project.DeploymentDataManager, bool)
	getDataManagerMutex       sync.RWMutex
	getDataManagerArgsForCall []struct {
	}
	getDataManagerReturns struct {
		result1 *project.DeploymentDataManager
		result2 bool
	}
	getDataManagerReturnsOnCall map[int]struct {
		result1 *project.DeploymentDataManager
		result2 bool
	}
	GetStatusStub        func(*client.Client) (common.DeploymentStatus, error)
	getStatusMutex       sync.RWMutex
	getStatusArgsForCall []struct {
		arg1 *client.Client
	}
	getStatusReturns struct {
		result1 common.DeploymentStatus
		result2 error
	}
	getStatusReturnsOnCall map[int]struct {
		result1 common.DeploymentStatus
		result2 error
	}
	InitializeStub        func(project.DeploymentConfig, io.Writer) error
	initializeMutex       sync.RWMutex
	initializeArgsForCall []struct {
		arg1 project.DeploymentConfig
		arg2 io.Writer
	}
	initializeReturns struct {
		result1 error
	}
	initializeReturnsOnCall map[int]struct {
		result1 error
	}
	PruneStub        func(*client.Client, io.Writer) error
	pruneMutex       sync.RWMutex
	pruneArgsForCall []struct {
		arg1 *client.Client
		arg2 io.Writer
	}
	pruneReturns struct {
		result1 error
	}
	pruneReturnsOnCall map[int]struct {
		result1 error
	}
	SetConfigStub        func(project.DeploymentConfig)
	setConfigMutex       sync.RWMutex
	setConfigArgsForCall []struct {
		arg1 project.DeploymentConfig
	}
	WatchStub        func(*client.Client) (<-chan string, <-chan error)
	watchMutex       sync.RWMutex
	watchArgsForCall []struct {
		arg1 *client.Client
	}
	watchReturns struct {
		result1 <-chan string
		result2 <-chan error
	}
	watchReturnsOnCall map[int]struct {
		result1 <-chan string
		result2 <-chan error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDeployer) CompareRemotes(arg1 string) error {
	fake.compareRemotesMutex.Lock()
	ret, specificReturn := fake.compareRemotesReturnsOnCall[len(fake.compareRemotesArgsForCall)]
	fake.compareRemotesArgsForCall = append(fake.compareRemotesArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("CompareRemotes", []interface{}{arg1})
	fake.compareRemotesMutex.Unlock()
	if fake.CompareRemotesStub != nil {
		return fake.CompareRemotesStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.compareRemotesReturns
	return fakeReturns.result1
}

func (fake *FakeDeployer) CompareRemotesCallCount() int {
	fake.compareRemotesMutex.RLock()
	defer fake.compareRemotesMutex.RUnlock()
	return len(fake.compareRemotesArgsForCall)
}

func (fake *FakeDeployer) CompareRemotesCalls(stub func(string) error) {
	fake.compareRemotesMutex.Lock()
	defer fake.compareRemotesMutex.Unlock()
	fake.CompareRemotesStub = stub
}

func (fake *FakeDeployer) CompareRemotesArgsForCall(i int) string {
	fake.compareRemotesMutex.RLock()
	defer fake.compareRemotesMutex.RUnlock()
	argsForCall := fake.compareRemotesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeDeployer) CompareRemotesReturns(result1 error) {
	fake.compareRemotesMutex.Lock()
	defer fake.compareRemotesMutex.Unlock()
	fake.CompareRemotesStub = nil
	fake.compareRemotesReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeployer) CompareRemotesReturnsOnCall(i int, result1 error) {
	fake.compareRemotesMutex.Lock()
	defer fake.compareRemotesMutex.Unlock()
	fake.CompareRemotesStub = nil
	if fake.compareRemotesReturnsOnCall == nil {
		fake.compareRemotesReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.compareRemotesReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeployer) Deploy(arg1 *client.Client, arg2 io.Writer, arg3 project.DeployOptions) (func() error, error) {
	fake.deployMutex.Lock()
	ret, specificReturn := fake.deployReturnsOnCall[len(fake.deployArgsForCall)]
	fake.deployArgsForCall = append(fake.deployArgsForCall, struct {
		arg1 *client.Client
		arg2 io.Writer
		arg3 project.DeployOptions
	}{arg1, arg2, arg3})
	fake.recordInvocation("Deploy", []interface{}{arg1, arg2, arg3})
	fake.deployMutex.Unlock()
	if fake.DeployStub != nil {
		return fake.DeployStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.deployReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDeployer) DeployCallCount() int {
	fake.deployMutex.RLock()
	defer fake.deployMutex.RUnlock()
	return len(fake.deployArgsForCall)
}

func (fake *FakeDeployer) DeployCalls(stub func(*client.Client, io.Writer, project.DeployOptions) (func() error, error)) {
	fake.deployMutex.Lock()
	defer fake.deployMutex.Unlock()
	fake.DeployStub = stub
}

func (fake *FakeDeployer) DeployArgsForCall(i int) (*client.Client, io.Writer, project.DeployOptions) {
	fake.deployMutex.RLock()
	defer fake.deployMutex.RUnlock()
	argsForCall := fake.deployArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeDeployer) DeployReturns(result1 func() error, result2 error) {
	fake.deployMutex.Lock()
	defer fake.deployMutex.Unlock()
	fake.DeployStub = nil
	fake.deployReturns = struct {
		result1 func() error
		result2 error
	}{result1, result2}
}

func (fake *FakeDeployer) DeployReturnsOnCall(i int, result1 func() error, result2 error) {
	fake.deployMutex.Lock()
	defer fake.deployMutex.Unlock()
	fake.DeployStub = nil
	if fake.deployReturnsOnCall == nil {
		fake.deployReturnsOnCall = make(map[int]struct {
			result1 func() error
			result2 error
		})
	}
	fake.deployReturnsOnCall[i] = struct {
		result1 func() error
		result2 error
	}{result1, result2}
}

func (fake *FakeDeployer) Destroy(arg1 *client.Client, arg2 io.Writer) error {
	fake.destroyMutex.Lock()
	ret, specificReturn := fake.destroyReturnsOnCall[len(fake.destroyArgsForCall)]
	fake.destroyArgsForCall = append(fake.destroyArgsForCall, struct {
		arg1 *client.Client
		arg2 io.Writer
	}{arg1, arg2})
	fake.recordInvocation("Destroy", []interface{}{arg1, arg2})
	fake.destroyMutex.Unlock()
	if fake.DestroyStub != nil {
		return fake.DestroyStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.destroyReturns
	return fakeReturns.result1
}

func (fake *FakeDeployer) DestroyCallCount() int {
	fake.destroyMutex.RLock()
	defer fake.destroyMutex.RUnlock()
	return len(fake.destroyArgsForCall)
}

func (fake *FakeDeployer) DestroyCalls(stub func(*client.Client, io.Writer) error) {
	fake.destroyMutex.Lock()
	defer fake.destroyMutex.Unlock()
	fake.DestroyStub = stub
}

func (fake *FakeDeployer) DestroyArgsForCall(i int) (*client.Client, io.Writer) {
	fake.destroyMutex.RLock()
	defer fake.destroyMutex.RUnlock()
	argsForCall := fake.destroyArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDeployer) DestroyReturns(result1 error) {
	fake.destroyMutex.Lock()
	defer fake.destroyMutex.Unlock()
	fake.DestroyStub = nil
	fake.destroyReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeployer) DestroyReturnsOnCall(i int, result1 error) {
	fake.destroyMutex.Lock()
	defer fake.destroyMutex.Unlock()
	fake.DestroyStub = nil
	if fake.destroyReturnsOnCall == nil {
		fake.destroyReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.destroyReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeployer) Down(arg1 *client.Client, arg2 io.Writer) error {
	fake.downMutex.Lock()
	ret, specificReturn := fake.downReturnsOnCall[len(fake.downArgsForCall)]
	fake.downArgsForCall = append(fake.downArgsForCall, struct {
		arg1 *client.Client
		arg2 io.Writer
	}{arg1, arg2})
	fake.recordInvocation("Down", []interface{}{arg1, arg2})
	fake.downMutex.Unlock()
	if fake.DownStub != nil {
		return fake.DownStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.downReturns
	return fakeReturns.result1
}

func (fake *FakeDeployer) DownCallCount() int {
	fake.downMutex.RLock()
	defer fake.downMutex.RUnlock()
	return len(fake.downArgsForCall)
}

func (fake *FakeDeployer) DownCalls(stub func(*client.Client, io.Writer) error) {
	fake.downMutex.Lock()
	defer fake.downMutex.Unlock()
	fake.DownStub = stub
}

func (fake *FakeDeployer) DownArgsForCall(i int) (*client.Client, io.Writer) {
	fake.downMutex.RLock()
	defer fake.downMutex.RUnlock()
	argsForCall := fake.downArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDeployer) DownReturns(result1 error) {
	fake.downMutex.Lock()
	defer fake.downMutex.Unlock()
	fake.DownStub = nil
	fake.downReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeployer) DownReturnsOnCall(i int, result1 error) {
	fake.downMutex.Lock()
	defer fake.downMutex.Unlock()
	fake.DownStub = nil
	if fake.downReturnsOnCall == nil {
		fake.downReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.downReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeployer) GetBranch() string {
	fake.getBranchMutex.Lock()
	ret, specificReturn := fake.getBranchReturnsOnCall[len(fake.getBranchArgsForCall)]
	fake.getBranchArgsForCall = append(fake.getBranchArgsForCall, struct {
	}{})
	fake.recordInvocation("GetBranch", []interface{}{})
	fake.getBranchMutex.Unlock()
	if fake.GetBranchStub != nil {
		return fake.GetBranchStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getBranchReturns
	return fakeReturns.result1
}

func (fake *FakeDeployer) GetBranchCallCount() int {
	fake.getBranchMutex.RLock()
	defer fake.getBranchMutex.RUnlock()
	return len(fake.getBranchArgsForCall)
}

func (fake *FakeDeployer) GetBranchCalls(stub func() string) {
	fake.getBranchMutex.Lock()
	defer fake.getBranchMutex.Unlock()
	fake.GetBranchStub = stub
}

func (fake *FakeDeployer) GetBranchReturns(result1 string) {
	fake.getBranchMutex.Lock()
	defer fake.getBranchMutex.Unlock()
	fake.GetBranchStub = nil
	fake.getBranchReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeDeployer) GetBranchReturnsOnCall(i int, result1 string) {
	fake.getBranchMutex.Lock()
	defer fake.getBranchMutex.Unlock()
	fake.GetBranchStub = nil
	if fake.getBranchReturnsOnCall == nil {
		fake.getBranchReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.getBranchReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeDeployer) GetDataManager() (*project.DeploymentDataManager, bool) {
	fake.getDataManagerMutex.Lock()
	ret, specificReturn := fake.getDataManagerReturnsOnCall[len(fake.getDataManagerArgsForCall)]
	fake.getDataManagerArgsForCall = append(fake.getDataManagerArgsForCall, struct {
	}{})
	fake.recordInvocation("GetDataManager", []interface{}{})
	fake.getDataManagerMutex.Unlock()
	if fake.GetDataManagerStub != nil {
		return fake.GetDataManagerStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getDataManagerReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDeployer) GetDataManagerCallCount() int {
	fake.getDataManagerMutex.RLock()
	defer fake.getDataManagerMutex.RUnlock()
	return len(fake.getDataManagerArgsForCall)
}

func (fake *FakeDeployer) GetDataManagerCalls(stub func() (*project.DeploymentDataManager, bool)) {
	fake.getDataManagerMutex.Lock()
	defer fake.getDataManagerMutex.Unlock()
	fake.GetDataManagerStub = stub
}

func (fake *FakeDeployer) GetDataManagerReturns(result1 *project.DeploymentDataManager, result2 bool) {
	fake.getDataManagerMutex.Lock()
	defer fake.getDataManagerMutex.Unlock()
	fake.GetDataManagerStub = nil
	fake.getDataManagerReturns = struct {
		result1 *project.DeploymentDataManager
		result2 bool
	}{result1, result2}
}

func (fake *FakeDeployer) GetDataManagerReturnsOnCall(i int, result1 *project.DeploymentDataManager, result2 bool) {
	fake.getDataManagerMutex.Lock()
	defer fake.getDataManagerMutex.Unlock()
	fake.GetDataManagerStub = nil
	if fake.getDataManagerReturnsOnCall == nil {
		fake.getDataManagerReturnsOnCall = make(map[int]struct {
			result1 *project.DeploymentDataManager
			result2 bool
		})
	}
	fake.getDataManagerReturnsOnCall[i] = struct {
		result1 *project.DeploymentDataManager
		result2 bool
	}{result1, result2}
}

func (fake *FakeDeployer) GetStatus(arg1 *client.Client) (common.DeploymentStatus, error) {
	fake.getStatusMutex.Lock()
	ret, specificReturn := fake.getStatusReturnsOnCall[len(fake.getStatusArgsForCall)]
	fake.getStatusArgsForCall = append(fake.getStatusArgsForCall, struct {
		arg1 *client.Client
	}{arg1})
	fake.recordInvocation("GetStatus", []interface{}{arg1})
	fake.getStatusMutex.Unlock()
	if fake.GetStatusStub != nil {
		return fake.GetStatusStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getStatusReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDeployer) GetStatusCallCount() int {
	fake.getStatusMutex.RLock()
	defer fake.getStatusMutex.RUnlock()
	return len(fake.getStatusArgsForCall)
}

func (fake *FakeDeployer) GetStatusCalls(stub func(*client.Client) (common.DeploymentStatus, error)) {
	fake.getStatusMutex.Lock()
	defer fake.getStatusMutex.Unlock()
	fake.GetStatusStub = stub
}

func (fake *FakeDeployer) GetStatusArgsForCall(i int) *client.Client {
	fake.getStatusMutex.RLock()
	defer fake.getStatusMutex.RUnlock()
	argsForCall := fake.getStatusArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeDeployer) GetStatusReturns(result1 common.DeploymentStatus, result2 error) {
	fake.getStatusMutex.Lock()
	defer fake.getStatusMutex.Unlock()
	fake.GetStatusStub = nil
	fake.getStatusReturns = struct {
		result1 common.DeploymentStatus
		result2 error
	}{result1, result2}
}

func (fake *FakeDeployer) GetStatusReturnsOnCall(i int, result1 common.DeploymentStatus, result2 error) {
	fake.getStatusMutex.Lock()
	defer fake.getStatusMutex.Unlock()
	fake.GetStatusStub = nil
	if fake.getStatusReturnsOnCall == nil {
		fake.getStatusReturnsOnCall = make(map[int]struct {
			result1 common.DeploymentStatus
			result2 error
		})
	}
	fake.getStatusReturnsOnCall[i] = struct {
		result1 common.DeploymentStatus
		result2 error
	}{result1, result2}
}

func (fake *FakeDeployer) Initialize(arg1 project.DeploymentConfig, arg2 io.Writer) error {
	fake.initializeMutex.Lock()
	ret, specificReturn := fake.initializeReturnsOnCall[len(fake.initializeArgsForCall)]
	fake.initializeArgsForCall = append(fake.initializeArgsForCall, struct {
		arg1 project.DeploymentConfig
		arg2 io.Writer
	}{arg1, arg2})
	fake.recordInvocation("Initialize", []interface{}{arg1, arg2})
	fake.initializeMutex.Unlock()
	if fake.InitializeStub != nil {
		return fake.InitializeStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.initializeReturns
	return fakeReturns.result1
}

func (fake *FakeDeployer) InitializeCallCount() int {
	fake.initializeMutex.RLock()
	defer fake.initializeMutex.RUnlock()
	return len(fake.initializeArgsForCall)
}

func (fake *FakeDeployer) InitializeCalls(stub func(project.DeploymentConfig, io.Writer) error) {
	fake.initializeMutex.Lock()
	defer fake.initializeMutex.Unlock()
	fake.InitializeStub = stub
}

func (fake *FakeDeployer) InitializeArgsForCall(i int) (project.DeploymentConfig, io.Writer) {
	fake.initializeMutex.RLock()
	defer fake.initializeMutex.RUnlock()
	argsForCall := fake.initializeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDeployer) InitializeReturns(result1 error) {
	fake.initializeMutex.Lock()
	defer fake.initializeMutex.Unlock()
	fake.InitializeStub = nil
	fake.initializeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeployer) InitializeReturnsOnCall(i int, result1 error) {
	fake.initializeMutex.Lock()
	defer fake.initializeMutex.Unlock()
	fake.InitializeStub = nil
	if fake.initializeReturnsOnCall == nil {
		fake.initializeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.initializeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeployer) Prune(arg1 *client.Client, arg2 io.Writer) error {
	fake.pruneMutex.Lock()
	ret, specificReturn := fake.pruneReturnsOnCall[len(fake.pruneArgsForCall)]
	fake.pruneArgsForCall = append(fake.pruneArgsForCall, struct {
		arg1 *client.Client
		arg2 io.Writer
	}{arg1, arg2})
	fake.recordInvocation("Prune", []interface{}{arg1, arg2})
	fake.pruneMutex.Unlock()
	if fake.PruneStub != nil {
		return fake.PruneStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.pruneReturns
	return fakeReturns.result1
}

func (fake *FakeDeployer) PruneCallCount() int {
	fake.pruneMutex.RLock()
	defer fake.pruneMutex.RUnlock()
	return len(fake.pruneArgsForCall)
}

func (fake *FakeDeployer) PruneCalls(stub func(*client.Client, io.Writer) error) {
	fake.pruneMutex.Lock()
	defer fake.pruneMutex.Unlock()
	fake.PruneStub = stub
}

func (fake *FakeDeployer) PruneArgsForCall(i int) (*client.Client, io.Writer) {
	fake.pruneMutex.RLock()
	defer fake.pruneMutex.RUnlock()
	argsForCall := fake.pruneArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDeployer) PruneReturns(result1 error) {
	fake.pruneMutex.Lock()
	defer fake.pruneMutex.Unlock()
	fake.PruneStub = nil
	fake.pruneReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeployer) PruneReturnsOnCall(i int, result1 error) {
	fake.pruneMutex.Lock()
	defer fake.pruneMutex.Unlock()
	fake.PruneStub = nil
	if fake.pruneReturnsOnCall == nil {
		fake.pruneReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.pruneReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeDeployer) SetConfig(arg1 project.DeploymentConfig) {
	fake.setConfigMutex.Lock()
	fake.setConfigArgsForCall = append(fake.setConfigArgsForCall, struct {
		arg1 project.DeploymentConfig
	}{arg1})
	fake.recordInvocation("SetConfig", []interface{}{arg1})
	fake.setConfigMutex.Unlock()
	if fake.SetConfigStub != nil {
		fake.SetConfigStub(arg1)
	}
}

func (fake *FakeDeployer) SetConfigCallCount() int {
	fake.setConfigMutex.RLock()
	defer fake.setConfigMutex.RUnlock()
	return len(fake.setConfigArgsForCall)
}

func (fake *FakeDeployer) SetConfigCalls(stub func(project.DeploymentConfig)) {
	fake.setConfigMutex.Lock()
	defer fake.setConfigMutex.Unlock()
	fake.SetConfigStub = stub
}

func (fake *FakeDeployer) SetConfigArgsForCall(i int) project.DeploymentConfig {
	fake.setConfigMutex.RLock()
	defer fake.setConfigMutex.RUnlock()
	argsForCall := fake.setConfigArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeDeployer) Watch(arg1 *client.Client) (<-chan string, <-chan error) {
	fake.watchMutex.Lock()
	ret, specificReturn := fake.watchReturnsOnCall[len(fake.watchArgsForCall)]
	fake.watchArgsForCall = append(fake.watchArgsForCall, struct {
		arg1 *client.Client
	}{arg1})
	fake.recordInvocation("Watch", []interface{}{arg1})
	fake.watchMutex.Unlock()
	if fake.WatchStub != nil {
		return fake.WatchStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.watchReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDeployer) WatchCallCount() int {
	fake.watchMutex.RLock()
	defer fake.watchMutex.RUnlock()
	return len(fake.watchArgsForCall)
}

func (fake *FakeDeployer) WatchCalls(stub func(*client.Client) (<-chan string, <-chan error)) {
	fake.watchMutex.Lock()
	defer fake.watchMutex.Unlock()
	fake.WatchStub = stub
}

func (fake *FakeDeployer) WatchArgsForCall(i int) *client.Client {
	fake.watchMutex.RLock()
	defer fake.watchMutex.RUnlock()
	argsForCall := fake.watchArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeDeployer) WatchReturns(result1 <-chan string, result2 <-chan error) {
	fake.watchMutex.Lock()
	defer fake.watchMutex.Unlock()
	fake.WatchStub = nil
	fake.watchReturns = struct {
		result1 <-chan string
		result2 <-chan error
	}{result1, result2}
}

func (fake *FakeDeployer) WatchReturnsOnCall(i int, result1 <-chan string, result2 <-chan error) {
	fake.watchMutex.Lock()
	defer fake.watchMutex.Unlock()
	fake.WatchStub = nil
	if fake.watchReturnsOnCall == nil {
		fake.watchReturnsOnCall = make(map[int]struct {
			result1 <-chan string
			result2 <-chan error
		})
	}
	fake.watchReturnsOnCall[i] = struct {
		result1 <-chan string
		result2 <-chan error
	}{result1, result2}
}

func (fake *FakeDeployer) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.compareRemotesMutex.RLock()
	defer fake.compareRemotesMutex.RUnlock()
	fake.deployMutex.RLock()
	defer fake.deployMutex.RUnlock()
	fake.destroyMutex.RLock()
	defer fake.destroyMutex.RUnlock()
	fake.downMutex.RLock()
	defer fake.downMutex.RUnlock()
	fake.getBranchMutex.RLock()
	defer fake.getBranchMutex.RUnlock()
	fake.getDataManagerMutex.RLock()
	defer fake.getDataManagerMutex.RUnlock()
	fake.getStatusMutex.RLock()
	defer fake.getStatusMutex.RUnlock()
	fake.initializeMutex.RLock()
	defer fake.initializeMutex.RUnlock()
	fake.pruneMutex.RLock()
	defer fake.pruneMutex.RUnlock()
	fake.setConfigMutex.RLock()
	defer fake.setConfigMutex.RUnlock()
	fake.watchMutex.RLock()
	defer fake.watchMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeDeployer) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ project.Deployer = new(FakeDeployer)
