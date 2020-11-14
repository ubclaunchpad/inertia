// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/notify"
)

type FakeNotifier struct {
	IsEqualStub        func(notify.Notifier) bool
	isEqualMutex       sync.RWMutex
	isEqualArgsForCall []struct {
		arg1 notify.Notifier
	}
	isEqualReturns struct {
		result1 bool
	}
	isEqualReturnsOnCall map[int]struct {
		result1 bool
	}
	NotifyStub        func(string, notify.Options) error
	notifyMutex       sync.RWMutex
	notifyArgsForCall []struct {
		arg1 string
		arg2 notify.Options
	}
	notifyReturns struct {
		result1 error
	}
	notifyReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeNotifier) IsEqual(arg1 notify.Notifier) bool {
	fake.isEqualMutex.Lock()
	ret, specificReturn := fake.isEqualReturnsOnCall[len(fake.isEqualArgsForCall)]
	fake.isEqualArgsForCall = append(fake.isEqualArgsForCall, struct {
		arg1 notify.Notifier
	}{arg1})
	stub := fake.IsEqualStub
	fakeReturns := fake.isEqualReturns
	fake.recordInvocation("IsEqual", []interface{}{arg1})
	fake.isEqualMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeNotifier) IsEqualCallCount() int {
	fake.isEqualMutex.RLock()
	defer fake.isEqualMutex.RUnlock()
	return len(fake.isEqualArgsForCall)
}

func (fake *FakeNotifier) IsEqualCalls(stub func(notify.Notifier) bool) {
	fake.isEqualMutex.Lock()
	defer fake.isEqualMutex.Unlock()
	fake.IsEqualStub = stub
}

func (fake *FakeNotifier) IsEqualArgsForCall(i int) notify.Notifier {
	fake.isEqualMutex.RLock()
	defer fake.isEqualMutex.RUnlock()
	argsForCall := fake.isEqualArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeNotifier) IsEqualReturns(result1 bool) {
	fake.isEqualMutex.Lock()
	defer fake.isEqualMutex.Unlock()
	fake.IsEqualStub = nil
	fake.isEqualReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeNotifier) IsEqualReturnsOnCall(i int, result1 bool) {
	fake.isEqualMutex.Lock()
	defer fake.isEqualMutex.Unlock()
	fake.IsEqualStub = nil
	if fake.isEqualReturnsOnCall == nil {
		fake.isEqualReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.isEqualReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *FakeNotifier) Notify(arg1 string, arg2 notify.Options) error {
	fake.notifyMutex.Lock()
	ret, specificReturn := fake.notifyReturnsOnCall[len(fake.notifyArgsForCall)]
	fake.notifyArgsForCall = append(fake.notifyArgsForCall, struct {
		arg1 string
		arg2 notify.Options
	}{arg1, arg2})
	stub := fake.NotifyStub
	fakeReturns := fake.notifyReturns
	fake.recordInvocation("Notify", []interface{}{arg1, arg2})
	fake.notifyMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeNotifier) NotifyCallCount() int {
	fake.notifyMutex.RLock()
	defer fake.notifyMutex.RUnlock()
	return len(fake.notifyArgsForCall)
}

func (fake *FakeNotifier) NotifyCalls(stub func(string, notify.Options) error) {
	fake.notifyMutex.Lock()
	defer fake.notifyMutex.Unlock()
	fake.NotifyStub = stub
}

func (fake *FakeNotifier) NotifyArgsForCall(i int) (string, notify.Options) {
	fake.notifyMutex.RLock()
	defer fake.notifyMutex.RUnlock()
	argsForCall := fake.notifyArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeNotifier) NotifyReturns(result1 error) {
	fake.notifyMutex.Lock()
	defer fake.notifyMutex.Unlock()
	fake.NotifyStub = nil
	fake.notifyReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeNotifier) NotifyReturnsOnCall(i int, result1 error) {
	fake.notifyMutex.Lock()
	defer fake.notifyMutex.Unlock()
	fake.NotifyStub = nil
	if fake.notifyReturnsOnCall == nil {
		fake.notifyReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.notifyReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeNotifier) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.isEqualMutex.RLock()
	defer fake.isEqualMutex.RUnlock()
	fake.notifyMutex.RLock()
	defer fake.notifyMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeNotifier) recordInvocation(key string, args []interface{}) {
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

var _ notify.Notifier = new(FakeNotifier)
