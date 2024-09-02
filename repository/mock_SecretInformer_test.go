// Code generated by mockery v2.42.1. DO NOT EDIT.

package repository

import (
	mock "github.com/stretchr/testify/mock"
	cache "k8s.io/client-go/tools/cache"

	v1 "k8s.io/client-go/listers/core/v1"
)

// MockSecretInformer is an autogenerated mock type for the SecretInformer type
type MockSecretInformer struct {
	mock.Mock
}

type MockSecretInformer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockSecretInformer) EXPECT() *MockSecretInformer_Expecter {
	return &MockSecretInformer_Expecter{mock: &_m.Mock}
}

// Informer provides a mock function with given fields:
func (_m *MockSecretInformer) Informer() cache.SharedIndexInformer {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Informer")
	}

	var r0 cache.SharedIndexInformer
	if rf, ok := ret.Get(0).(func() cache.SharedIndexInformer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cache.SharedIndexInformer)
		}
	}

	return r0
}

// MockSecretInformer_Informer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Informer'
type MockSecretInformer_Informer_Call struct {
	*mock.Call
}

// Informer is a helper method to define mock.On call
func (_e *MockSecretInformer_Expecter) Informer() *MockSecretInformer_Informer_Call {
	return &MockSecretInformer_Informer_Call{Call: _e.mock.On("Informer")}
}

func (_c *MockSecretInformer_Informer_Call) Run(run func()) *MockSecretInformer_Informer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockSecretInformer_Informer_Call) Return(_a0 cache.SharedIndexInformer) *MockSecretInformer_Informer_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockSecretInformer_Informer_Call) RunAndReturn(run func() cache.SharedIndexInformer) *MockSecretInformer_Informer_Call {
	_c.Call.Return(run)
	return _c
}

// Lister provides a mock function with given fields:
func (_m *MockSecretInformer) Lister() v1.SecretLister {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Lister")
	}

	var r0 v1.SecretLister
	if rf, ok := ret.Get(0).(func() v1.SecretLister); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1.SecretLister)
		}
	}

	return r0
}

// MockSecretInformer_Lister_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Lister'
type MockSecretInformer_Lister_Call struct {
	*mock.Call
}

// Lister is a helper method to define mock.On call
func (_e *MockSecretInformer_Expecter) Lister() *MockSecretInformer_Lister_Call {
	return &MockSecretInformer_Lister_Call{Call: _e.mock.On("Lister")}
}

func (_c *MockSecretInformer_Lister_Call) Run(run func()) *MockSecretInformer_Lister_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockSecretInformer_Lister_Call) Return(_a0 v1.SecretLister) *MockSecretInformer_Lister_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockSecretInformer_Lister_Call) RunAndReturn(run func() v1.SecretLister) *MockSecretInformer_Lister_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockSecretInformer creates a new instance of MockSecretInformer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockSecretInformer(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSecretInformer {
	mock := &MockSecretInformer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
