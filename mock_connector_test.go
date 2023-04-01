// Code generated by mockery v2.20.2. DO NOT EDIT.

package portforwarder

import (
	mock "github.com/stretchr/testify/mock"
	kubernetes "k8s.io/client-go/kubernetes"

	rest "k8s.io/client-go/rest"
)

// mockConnector is an autogenerated mock type for the connector type
type mockConnector struct {
	mock.Mock
}

type mockConnector_Expecter struct {
	mock *mock.Mock
}

func (_m *mockConnector) EXPECT() *mockConnector_Expecter {
	return &mockConnector_Expecter{mock: &_m.Mock}
}

// Connect provides a mock function with given fields:
func (_m *mockConnector) Connect() (*rest.Config, *kubernetes.Clientset, error) {
	ret := _m.Called()

	var r0 *rest.Config
	var r1 *kubernetes.Clientset
	var r2 error
	if rf, ok := ret.Get(0).(func() (*rest.Config, *kubernetes.Clientset, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *rest.Config); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rest.Config)
		}
	}

	if rf, ok := ret.Get(1).(func() *kubernetes.Clientset); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*kubernetes.Clientset)
		}
	}

	if rf, ok := ret.Get(2).(func() error); ok {
		r2 = rf()
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// mockConnector_Connect_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Connect'
type mockConnector_Connect_Call struct {
	*mock.Call
}

// Connect is a helper method to define mock.On call
func (_e *mockConnector_Expecter) Connect() *mockConnector_Connect_Call {
	return &mockConnector_Connect_Call{Call: _e.mock.On("Connect")}
}

func (_c *mockConnector_Connect_Call) Run(run func()) *mockConnector_Connect_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockConnector_Connect_Call) Return(_a0 *rest.Config, _a1 *kubernetes.Clientset, _a2 error) *mockConnector_Connect_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *mockConnector_Connect_Call) RunAndReturn(run func() (*rest.Config, *kubernetes.Clientset, error)) *mockConnector_Connect_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTnewMockConnector interface {
	mock.TestingT
	Cleanup(func())
}

// newMockConnector creates a new instance of mockConnector. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newMockConnector(t mockConstructorTestingTnewMockConnector) *mockConnector {
	mock := &mockConnector{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
