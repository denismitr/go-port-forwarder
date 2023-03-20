// Code generated by mockery v2.20.2. DO NOT EDIT.

package portforwarder

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

// mockPodProvider is an autogenerated mock type for the podProvider type
type mockPodProvider struct {
	mock.Mock
}

type mockPodProvider_Expecter struct {
	mock *mock.Mock
}

func (_m *mockPodProvider) EXPECT() *mockPodProvider_Expecter {
	return &mockPodProvider_Expecter{mock: &_m.Mock}
}

// listPods provides a mock function with given fields: _a0, _a1
func (_m *mockPodProvider) listPods(_a0 context.Context, _a1 *listPodsCommand) (*v1.PodList, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *v1.PodList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *listPodsCommand) (*v1.PodList, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *listPodsCommand) *v1.PodList); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.PodList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *listPodsCommand) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockPodProvider_listPods_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'listPods'
type mockPodProvider_listPods_Call struct {
	*mock.Call
}

// listPods is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *listPodsCommand
func (_e *mockPodProvider_Expecter) listPods(_a0 interface{}, _a1 interface{}) *mockPodProvider_listPods_Call {
	return &mockPodProvider_listPods_Call{Call: _e.mock.On("listPods", _a0, _a1)}
}

func (_c *mockPodProvider_listPods_Call) Run(run func(_a0 context.Context, _a1 *listPodsCommand)) *mockPodProvider_listPods_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*listPodsCommand))
	})
	return _c
}

func (_c *mockPodProvider_listPods_Call) Return(_a0 *v1.PodList, _a1 error) *mockPodProvider_listPods_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockPodProvider_listPods_Call) RunAndReturn(run func(context.Context, *listPodsCommand) (*v1.PodList, error)) *mockPodProvider_listPods_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTnewMockPodProvider interface {
	mock.TestingT
	Cleanup(func())
}

// newMockPodProvider creates a new instance of mockPodProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newMockPodProvider(t mockConstructorTestingTnewMockPodProvider) *mockPodProvider {
	mock := &mockPodProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}