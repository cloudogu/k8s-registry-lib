// Code generated by mockery v2.42.1. DO NOT EDIT.

package global

import (
	core "github.com/cloudogu/cesapp-lib/core"
	mock "github.com/stretchr/testify/mock"
)

// mockDoguRegistry is an autogenerated mock type for the doguRegistry type
type mockDoguRegistry struct {
	mock.Mock
}

type mockDoguRegistry_Expecter struct {
	mock *mock.Mock
}

func (_m *mockDoguRegistry) EXPECT() *mockDoguRegistry_Expecter {
	return &mockDoguRegistry_Expecter{mock: &_m.Mock}
}

// Enable provides a mock function with given fields: dogu
func (_m *mockDoguRegistry) Enable(dogu *core.Dogu) error {
	ret := _m.Called(dogu)

	if len(ret) == 0 {
		panic("no return value specified for Enable")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*core.Dogu) error); ok {
		r0 = rf(dogu)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockDoguRegistry_Enable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Enable'
type mockDoguRegistry_Enable_Call struct {
	*mock.Call
}

// Enable is a helper method to define mock.On call
//   - dogu *core.Dogu
func (_e *mockDoguRegistry_Expecter) Enable(dogu interface{}) *mockDoguRegistry_Enable_Call {
	return &mockDoguRegistry_Enable_Call{Call: _e.mock.On("Enable", dogu)}
}

func (_c *mockDoguRegistry_Enable_Call) Run(run func(dogu *core.Dogu)) *mockDoguRegistry_Enable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*core.Dogu))
	})
	return _c
}

func (_c *mockDoguRegistry_Enable_Call) Return(_a0 error) *mockDoguRegistry_Enable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDoguRegistry_Enable_Call) RunAndReturn(run func(*core.Dogu) error) *mockDoguRegistry_Enable_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: name
func (_m *mockDoguRegistry) Get(name string) (*core.Dogu, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *core.Dogu
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*core.Dogu, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) *core.Dogu); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.Dogu)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDoguRegistry_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type mockDoguRegistry_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - name string
func (_e *mockDoguRegistry_Expecter) Get(name interface{}) *mockDoguRegistry_Get_Call {
	return &mockDoguRegistry_Get_Call{Call: _e.mock.On("Get", name)}
}

func (_c *mockDoguRegistry_Get_Call) Run(run func(name string)) *mockDoguRegistry_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockDoguRegistry_Get_Call) Return(_a0 *core.Dogu, _a1 error) *mockDoguRegistry_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockDoguRegistry_Get_Call) RunAndReturn(run func(string) (*core.Dogu, error)) *mockDoguRegistry_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetAll provides a mock function with given fields:
func (_m *mockDoguRegistry) GetAll() ([]*core.Dogu, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetAll")
	}

	var r0 []*core.Dogu
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]*core.Dogu, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []*core.Dogu); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*core.Dogu)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDoguRegistry_GetAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAll'
type mockDoguRegistry_GetAll_Call struct {
	*mock.Call
}

// GetAll is a helper method to define mock.On call
func (_e *mockDoguRegistry_Expecter) GetAll() *mockDoguRegistry_GetAll_Call {
	return &mockDoguRegistry_GetAll_Call{Call: _e.mock.On("GetAll")}
}

func (_c *mockDoguRegistry_GetAll_Call) Run(run func()) *mockDoguRegistry_GetAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockDoguRegistry_GetAll_Call) Return(_a0 []*core.Dogu, _a1 error) *mockDoguRegistry_GetAll_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockDoguRegistry_GetAll_Call) RunAndReturn(run func() ([]*core.Dogu, error)) *mockDoguRegistry_GetAll_Call {
	_c.Call.Return(run)
	return _c
}

// IsEnabled provides a mock function with given fields: name
func (_m *mockDoguRegistry) IsEnabled(name string) (bool, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for IsEnabled")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (bool, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDoguRegistry_IsEnabled_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsEnabled'
type mockDoguRegistry_IsEnabled_Call struct {
	*mock.Call
}

// IsEnabled is a helper method to define mock.On call
//   - name string
func (_e *mockDoguRegistry_Expecter) IsEnabled(name interface{}) *mockDoguRegistry_IsEnabled_Call {
	return &mockDoguRegistry_IsEnabled_Call{Call: _e.mock.On("IsEnabled", name)}
}

func (_c *mockDoguRegistry_IsEnabled_Call) Run(run func(name string)) *mockDoguRegistry_IsEnabled_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockDoguRegistry_IsEnabled_Call) Return(_a0 bool, _a1 error) *mockDoguRegistry_IsEnabled_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockDoguRegistry_IsEnabled_Call) RunAndReturn(run func(string) (bool, error)) *mockDoguRegistry_IsEnabled_Call {
	_c.Call.Return(run)
	return _c
}

// Register provides a mock function with given fields: dogu
func (_m *mockDoguRegistry) Register(dogu *core.Dogu) error {
	ret := _m.Called(dogu)

	if len(ret) == 0 {
		panic("no return value specified for Register")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*core.Dogu) error); ok {
		r0 = rf(dogu)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockDoguRegistry_Register_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Register'
type mockDoguRegistry_Register_Call struct {
	*mock.Call
}

// Register is a helper method to define mock.On call
//   - dogu *core.Dogu
func (_e *mockDoguRegistry_Expecter) Register(dogu interface{}) *mockDoguRegistry_Register_Call {
	return &mockDoguRegistry_Register_Call{Call: _e.mock.On("Register", dogu)}
}

func (_c *mockDoguRegistry_Register_Call) Run(run func(dogu *core.Dogu)) *mockDoguRegistry_Register_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*core.Dogu))
	})
	return _c
}

func (_c *mockDoguRegistry_Register_Call) Return(_a0 error) *mockDoguRegistry_Register_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDoguRegistry_Register_Call) RunAndReturn(run func(*core.Dogu) error) *mockDoguRegistry_Register_Call {
	_c.Call.Return(run)
	return _c
}

// Unregister provides a mock function with given fields: name
func (_m *mockDoguRegistry) Unregister(name string) error {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Unregister")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockDoguRegistry_Unregister_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Unregister'
type mockDoguRegistry_Unregister_Call struct {
	*mock.Call
}

// Unregister is a helper method to define mock.On call
//   - name string
func (_e *mockDoguRegistry_Expecter) Unregister(name interface{}) *mockDoguRegistry_Unregister_Call {
	return &mockDoguRegistry_Unregister_Call{Call: _e.mock.On("Unregister", name)}
}

func (_c *mockDoguRegistry_Unregister_Call) Run(run func(name string)) *mockDoguRegistry_Unregister_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockDoguRegistry_Unregister_Call) Return(_a0 error) *mockDoguRegistry_Unregister_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockDoguRegistry_Unregister_Call) RunAndReturn(run func(string) error) *mockDoguRegistry_Unregister_Call {
	_c.Call.Return(run)
	return _c
}

// newMockDoguRegistry creates a new instance of mockDoguRegistry. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockDoguRegistry(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockDoguRegistry {
	mock := &mockDoguRegistry{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}