// Code generated by mockery v2.42.1. DO NOT EDIT.

package registry

import mock "github.com/stretchr/testify/mock"

// mockConfigData is an autogenerated mock type for the configData type
type mockConfigData struct {
	mock.Mock
}

type mockConfigData_Expecter struct {
	mock *mock.Mock
}

func (_m *mockConfigData) EXPECT() *mockConfigData_Expecter {
	return &mockConfigData_Expecter{mock: &_m.Mock}
}

// get provides a mock function with given fields:
func (_m *mockConfigData) get() map[string]string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for get")
	}

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func() map[string]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	return r0
}

// mockConfigData_get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'get'
type mockConfigData_get_Call struct {
	*mock.Call
}

// get is a helper method to define mock.On call
func (_e *mockConfigData_Expecter) get() *mockConfigData_get_Call {
	return &mockConfigData_get_Call{Call: _e.mock.On("get")}
}

func (_c *mockConfigData_get_Call) Run(run func()) *mockConfigData_get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockConfigData_get_Call) Return(_a0 map[string]string) *mockConfigData_get_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockConfigData_get_Call) RunAndReturn(run func() map[string]string) *mockConfigData_get_Call {
	_c.Call.Return(run)
	return _c
}

// newMockConfigData creates a new instance of mockConfigData. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockConfigData(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockConfigData {
	mock := &mockConfigData{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
