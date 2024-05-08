// Code generated by mockery v2.42.1. DO NOT EDIT.

package local

import (
	registry "github.com/cloudogu/cesapp-lib/registry"
	mock "github.com/stretchr/testify/mock"
)

// mockEtcdRegistry is an autogenerated mock type for the etcdRegistry type
type mockEtcdRegistry struct {
	mock.Mock
}

type mockEtcdRegistry_Expecter struct {
	mock *mock.Mock
}

func (_m *mockEtcdRegistry) EXPECT() *mockEtcdRegistry_Expecter {
	return &mockEtcdRegistry_Expecter{mock: &_m.Mock}
}

// BlueprintRegistry provides a mock function with given fields:
func (_m *mockEtcdRegistry) BlueprintRegistry() registry.ConfigurationContext {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for BlueprintRegistry")
	}

	var r0 registry.ConfigurationContext
	if rf, ok := ret.Get(0).(func() registry.ConfigurationContext); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(registry.ConfigurationContext)
		}
	}

	return r0
}

// mockEtcdRegistry_BlueprintRegistry_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BlueprintRegistry'
type mockEtcdRegistry_BlueprintRegistry_Call struct {
	*mock.Call
}

// BlueprintRegistry is a helper method to define mock.On call
func (_e *mockEtcdRegistry_Expecter) BlueprintRegistry() *mockEtcdRegistry_BlueprintRegistry_Call {
	return &mockEtcdRegistry_BlueprintRegistry_Call{Call: _e.mock.On("BlueprintRegistry")}
}

func (_c *mockEtcdRegistry_BlueprintRegistry_Call) Run(run func()) *mockEtcdRegistry_BlueprintRegistry_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockEtcdRegistry_BlueprintRegistry_Call) Return(_a0 registry.ConfigurationContext) *mockEtcdRegistry_BlueprintRegistry_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdRegistry_BlueprintRegistry_Call) RunAndReturn(run func() registry.ConfigurationContext) *mockEtcdRegistry_BlueprintRegistry_Call {
	_c.Call.Return(run)
	return _c
}

// DoguConfig provides a mock function with given fields: dogu
func (_m *mockEtcdRegistry) DoguConfig(dogu string) registry.ConfigurationContext {
	ret := _m.Called(dogu)

	if len(ret) == 0 {
		panic("no return value specified for DoguConfig")
	}

	var r0 registry.ConfigurationContext
	if rf, ok := ret.Get(0).(func(string) registry.ConfigurationContext); ok {
		r0 = rf(dogu)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(registry.ConfigurationContext)
		}
	}

	return r0
}

// mockEtcdRegistry_DoguConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DoguConfig'
type mockEtcdRegistry_DoguConfig_Call struct {
	*mock.Call
}

// DoguConfig is a helper method to define mock.On call
//   - dogu string
func (_e *mockEtcdRegistry_Expecter) DoguConfig(dogu interface{}) *mockEtcdRegistry_DoguConfig_Call {
	return &mockEtcdRegistry_DoguConfig_Call{Call: _e.mock.On("DoguConfig", dogu)}
}

func (_c *mockEtcdRegistry_DoguConfig_Call) Run(run func(dogu string)) *mockEtcdRegistry_DoguConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockEtcdRegistry_DoguConfig_Call) Return(_a0 registry.ConfigurationContext) *mockEtcdRegistry_DoguConfig_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdRegistry_DoguConfig_Call) RunAndReturn(run func(string) registry.ConfigurationContext) *mockEtcdRegistry_DoguConfig_Call {
	_c.Call.Return(run)
	return _c
}

// DoguRegistry provides a mock function with given fields:
func (_m *mockEtcdRegistry) DoguRegistry() registry.DoguRegistry {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for DoguRegistry")
	}

	var r0 registry.DoguRegistry
	if rf, ok := ret.Get(0).(func() registry.DoguRegistry); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(registry.DoguRegistry)
		}
	}

	return r0
}

// mockEtcdRegistry_DoguRegistry_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DoguRegistry'
type mockEtcdRegistry_DoguRegistry_Call struct {
	*mock.Call
}

// DoguRegistry is a helper method to define mock.On call
func (_e *mockEtcdRegistry_Expecter) DoguRegistry() *mockEtcdRegistry_DoguRegistry_Call {
	return &mockEtcdRegistry_DoguRegistry_Call{Call: _e.mock.On("DoguRegistry")}
}

func (_c *mockEtcdRegistry_DoguRegistry_Call) Run(run func()) *mockEtcdRegistry_DoguRegistry_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockEtcdRegistry_DoguRegistry_Call) Return(_a0 registry.DoguRegistry) *mockEtcdRegistry_DoguRegistry_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdRegistry_DoguRegistry_Call) RunAndReturn(run func() registry.DoguRegistry) *mockEtcdRegistry_DoguRegistry_Call {
	_c.Call.Return(run)
	return _c
}

// GetNode provides a mock function with given fields:
func (_m *mockEtcdRegistry) GetNode() (registry.Node, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetNode")
	}

	var r0 registry.Node
	var r1 error
	if rf, ok := ret.Get(0).(func() (registry.Node, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() registry.Node); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(registry.Node)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockEtcdRegistry_GetNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNode'
type mockEtcdRegistry_GetNode_Call struct {
	*mock.Call
}

// GetNode is a helper method to define mock.On call
func (_e *mockEtcdRegistry_Expecter) GetNode() *mockEtcdRegistry_GetNode_Call {
	return &mockEtcdRegistry_GetNode_Call{Call: _e.mock.On("GetNode")}
}

func (_c *mockEtcdRegistry_GetNode_Call) Run(run func()) *mockEtcdRegistry_GetNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockEtcdRegistry_GetNode_Call) Return(_a0 registry.Node, _a1 error) *mockEtcdRegistry_GetNode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockEtcdRegistry_GetNode_Call) RunAndReturn(run func() (registry.Node, error)) *mockEtcdRegistry_GetNode_Call {
	_c.Call.Return(run)
	return _c
}

// GlobalConfig provides a mock function with given fields:
func (_m *mockEtcdRegistry) GlobalConfig() registry.ConfigurationContext {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GlobalConfig")
	}

	var r0 registry.ConfigurationContext
	if rf, ok := ret.Get(0).(func() registry.ConfigurationContext); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(registry.ConfigurationContext)
		}
	}

	return r0
}

// mockEtcdRegistry_GlobalConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GlobalConfig'
type mockEtcdRegistry_GlobalConfig_Call struct {
	*mock.Call
}

// GlobalConfig is a helper method to define mock.On call
func (_e *mockEtcdRegistry_Expecter) GlobalConfig() *mockEtcdRegistry_GlobalConfig_Call {
	return &mockEtcdRegistry_GlobalConfig_Call{Call: _e.mock.On("GlobalConfig")}
}

func (_c *mockEtcdRegistry_GlobalConfig_Call) Run(run func()) *mockEtcdRegistry_GlobalConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockEtcdRegistry_GlobalConfig_Call) Return(_a0 registry.ConfigurationContext) *mockEtcdRegistry_GlobalConfig_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdRegistry_GlobalConfig_Call) RunAndReturn(run func() registry.ConfigurationContext) *mockEtcdRegistry_GlobalConfig_Call {
	_c.Call.Return(run)
	return _c
}

// HostConfig provides a mock function with given fields: hostService
func (_m *mockEtcdRegistry) HostConfig(hostService string) registry.ConfigurationContext {
	ret := _m.Called(hostService)

	if len(ret) == 0 {
		panic("no return value specified for HostConfig")
	}

	var r0 registry.ConfigurationContext
	if rf, ok := ret.Get(0).(func(string) registry.ConfigurationContext); ok {
		r0 = rf(hostService)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(registry.ConfigurationContext)
		}
	}

	return r0
}

// mockEtcdRegistry_HostConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HostConfig'
type mockEtcdRegistry_HostConfig_Call struct {
	*mock.Call
}

// HostConfig is a helper method to define mock.On call
//   - hostService string
func (_e *mockEtcdRegistry_Expecter) HostConfig(hostService interface{}) *mockEtcdRegistry_HostConfig_Call {
	return &mockEtcdRegistry_HostConfig_Call{Call: _e.mock.On("HostConfig", hostService)}
}

func (_c *mockEtcdRegistry_HostConfig_Call) Run(run func(hostService string)) *mockEtcdRegistry_HostConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockEtcdRegistry_HostConfig_Call) Return(_a0 registry.ConfigurationContext) *mockEtcdRegistry_HostConfig_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdRegistry_HostConfig_Call) RunAndReturn(run func(string) registry.ConfigurationContext) *mockEtcdRegistry_HostConfig_Call {
	_c.Call.Return(run)
	return _c
}

// RootConfig provides a mock function with given fields:
func (_m *mockEtcdRegistry) RootConfig() registry.WatchConfigurationContext {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RootConfig")
	}

	var r0 registry.WatchConfigurationContext
	if rf, ok := ret.Get(0).(func() registry.WatchConfigurationContext); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(registry.WatchConfigurationContext)
		}
	}

	return r0
}

// mockEtcdRegistry_RootConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RootConfig'
type mockEtcdRegistry_RootConfig_Call struct {
	*mock.Call
}

// RootConfig is a helper method to define mock.On call
func (_e *mockEtcdRegistry_Expecter) RootConfig() *mockEtcdRegistry_RootConfig_Call {
	return &mockEtcdRegistry_RootConfig_Call{Call: _e.mock.On("RootConfig")}
}

func (_c *mockEtcdRegistry_RootConfig_Call) Run(run func()) *mockEtcdRegistry_RootConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockEtcdRegistry_RootConfig_Call) Return(_a0 registry.WatchConfigurationContext) *mockEtcdRegistry_RootConfig_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdRegistry_RootConfig_Call) RunAndReturn(run func() registry.WatchConfigurationContext) *mockEtcdRegistry_RootConfig_Call {
	_c.Call.Return(run)
	return _c
}

// State provides a mock function with given fields: dogu
func (_m *mockEtcdRegistry) State(dogu string) registry.State {
	ret := _m.Called(dogu)

	if len(ret) == 0 {
		panic("no return value specified for State")
	}

	var r0 registry.State
	if rf, ok := ret.Get(0).(func(string) registry.State); ok {
		r0 = rf(dogu)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(registry.State)
		}
	}

	return r0
}

// mockEtcdRegistry_State_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'State'
type mockEtcdRegistry_State_Call struct {
	*mock.Call
}

// State is a helper method to define mock.On call
//   - dogu string
func (_e *mockEtcdRegistry_Expecter) State(dogu interface{}) *mockEtcdRegistry_State_Call {
	return &mockEtcdRegistry_State_Call{Call: _e.mock.On("State", dogu)}
}

func (_c *mockEtcdRegistry_State_Call) Run(run func(dogu string)) *mockEtcdRegistry_State_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockEtcdRegistry_State_Call) Return(_a0 registry.State) *mockEtcdRegistry_State_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdRegistry_State_Call) RunAndReturn(run func(string) registry.State) *mockEtcdRegistry_State_Call {
	_c.Call.Return(run)
	return _c
}

// newMockEtcdRegistry creates a new instance of mockEtcdRegistry. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockEtcdRegistry(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockEtcdRegistry {
	mock := &mockEtcdRegistry{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
