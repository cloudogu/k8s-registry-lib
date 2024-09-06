// Code generated by mockery v2.42.1. DO NOT EDIT.

package dogu

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockDoguVersionRegistry is an autogenerated mock type for the DoguVersionRegistry type
type MockDoguVersionRegistry struct {
	mock.Mock
}

type MockDoguVersionRegistry_Expecter struct {
	mock *mock.Mock
}

func (_m *MockDoguVersionRegistry) EXPECT() *MockDoguVersionRegistry_Expecter {
	return &MockDoguVersionRegistry_Expecter{mock: &_m.Mock}
}

// Enable provides a mock function with given fields: _a0, _a1
func (_m *MockDoguVersionRegistry) Enable(_a0 context.Context, _a1 DoguVersion) error {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Enable")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, DoguVersion) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDoguVersionRegistry_Enable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Enable'
type MockDoguVersionRegistry_Enable_Call struct {
	*mock.Call
}

// Enable is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 DoguVersion
func (_e *MockDoguVersionRegistry_Expecter) Enable(_a0 interface{}, _a1 interface{}) *MockDoguVersionRegistry_Enable_Call {
	return &MockDoguVersionRegistry_Enable_Call{Call: _e.mock.On("Enable", _a0, _a1)}
}

func (_c *MockDoguVersionRegistry_Enable_Call) Run(run func(_a0 context.Context, _a1 DoguVersion)) *MockDoguVersionRegistry_Enable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(DoguVersion))
	})
	return _c
}

func (_c *MockDoguVersionRegistry_Enable_Call) Return(_a0 error) *MockDoguVersionRegistry_Enable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDoguVersionRegistry_Enable_Call) RunAndReturn(run func(context.Context, DoguVersion) error) *MockDoguVersionRegistry_Enable_Call {
	_c.Call.Return(run)
	return _c
}

// GetCurrent provides a mock function with given fields: _a0, _a1
func (_m *MockDoguVersionRegistry) GetCurrent(_a0 context.Context, _a1 SimpleDoguName) (DoguVersion, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetCurrent")
	}

	var r0 DoguVersion
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, SimpleDoguName) (DoguVersion, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, SimpleDoguName) DoguVersion); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(DoguVersion)
	}

	if rf, ok := ret.Get(1).(func(context.Context, SimpleDoguName) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDoguVersionRegistry_GetCurrent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCurrent'
type MockDoguVersionRegistry_GetCurrent_Call struct {
	*mock.Call
}

// GetCurrent is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 SimpleDoguName
func (_e *MockDoguVersionRegistry_Expecter) GetCurrent(_a0 interface{}, _a1 interface{}) *MockDoguVersionRegistry_GetCurrent_Call {
	return &MockDoguVersionRegistry_GetCurrent_Call{Call: _e.mock.On("GetCurrent", _a0, _a1)}
}

func (_c *MockDoguVersionRegistry_GetCurrent_Call) Run(run func(_a0 context.Context, _a1 SimpleDoguName)) *MockDoguVersionRegistry_GetCurrent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(SimpleDoguName))
	})
	return _c
}

func (_c *MockDoguVersionRegistry_GetCurrent_Call) Return(_a0 DoguVersion, _a1 error) *MockDoguVersionRegistry_GetCurrent_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDoguVersionRegistry_GetCurrent_Call) RunAndReturn(run func(context.Context, SimpleDoguName) (DoguVersion, error)) *MockDoguVersionRegistry_GetCurrent_Call {
	_c.Call.Return(run)
	return _c
}

// GetCurrentOfAll provides a mock function with given fields: _a0
func (_m *MockDoguVersionRegistry) GetCurrentOfAll(_a0 context.Context) ([]DoguVersion, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetCurrentOfAll")
	}

	var r0 []DoguVersion
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]DoguVersion, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []DoguVersion); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]DoguVersion)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDoguVersionRegistry_GetCurrentOfAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCurrentOfAll'
type MockDoguVersionRegistry_GetCurrentOfAll_Call struct {
	*mock.Call
}

// GetCurrentOfAll is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *MockDoguVersionRegistry_Expecter) GetCurrentOfAll(_a0 interface{}) *MockDoguVersionRegistry_GetCurrentOfAll_Call {
	return &MockDoguVersionRegistry_GetCurrentOfAll_Call{Call: _e.mock.On("GetCurrentOfAll", _a0)}
}

func (_c *MockDoguVersionRegistry_GetCurrentOfAll_Call) Run(run func(_a0 context.Context)) *MockDoguVersionRegistry_GetCurrentOfAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockDoguVersionRegistry_GetCurrentOfAll_Call) Return(_a0 []DoguVersion, _a1 error) *MockDoguVersionRegistry_GetCurrentOfAll_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDoguVersionRegistry_GetCurrentOfAll_Call) RunAndReturn(run func(context.Context) ([]DoguVersion, error)) *MockDoguVersionRegistry_GetCurrentOfAll_Call {
	_c.Call.Return(run)
	return _c
}

// IsEnabled provides a mock function with given fields: _a0, _a1
func (_m *MockDoguVersionRegistry) IsEnabled(_a0 context.Context, _a1 DoguVersion) (bool, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for IsEnabled")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, DoguVersion) (bool, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, DoguVersion) bool); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, DoguVersion) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDoguVersionRegistry_IsEnabled_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsEnabled'
type MockDoguVersionRegistry_IsEnabled_Call struct {
	*mock.Call
}

// IsEnabled is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 DoguVersion
func (_e *MockDoguVersionRegistry_Expecter) IsEnabled(_a0 interface{}, _a1 interface{}) *MockDoguVersionRegistry_IsEnabled_Call {
	return &MockDoguVersionRegistry_IsEnabled_Call{Call: _e.mock.On("IsEnabled", _a0, _a1)}
}

func (_c *MockDoguVersionRegistry_IsEnabled_Call) Run(run func(_a0 context.Context, _a1 DoguVersion)) *MockDoguVersionRegistry_IsEnabled_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(DoguVersion))
	})
	return _c
}

func (_c *MockDoguVersionRegistry_IsEnabled_Call) Return(_a0 bool, _a1 error) *MockDoguVersionRegistry_IsEnabled_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDoguVersionRegistry_IsEnabled_Call) RunAndReturn(run func(context.Context, DoguVersion) (bool, error)) *MockDoguVersionRegistry_IsEnabled_Call {
	_c.Call.Return(run)
	return _c
}

// WatchAllCurrent provides a mock function with given fields: _a0
func (_m *MockDoguVersionRegistry) WatchAllCurrent(_a0 context.Context) (<-chan CurrentVersionsWatchResult, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for WatchAllCurrent")
	}

	var r0 <-chan CurrentVersionsWatchResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (<-chan CurrentVersionsWatchResult, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) <-chan CurrentVersionsWatchResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan CurrentVersionsWatchResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDoguVersionRegistry_WatchAllCurrent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WatchAllCurrent'
type MockDoguVersionRegistry_WatchAllCurrent_Call struct {
	*mock.Call
}

// WatchAllCurrent is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *MockDoguVersionRegistry_Expecter) WatchAllCurrent(_a0 interface{}) *MockDoguVersionRegistry_WatchAllCurrent_Call {
	return &MockDoguVersionRegistry_WatchAllCurrent_Call{Call: _e.mock.On("WatchAllCurrent", _a0)}
}

func (_c *MockDoguVersionRegistry_WatchAllCurrent_Call) Run(run func(_a0 context.Context)) *MockDoguVersionRegistry_WatchAllCurrent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockDoguVersionRegistry_WatchAllCurrent_Call) Return(_a0 <-chan CurrentVersionsWatchResult, _a1 error) *MockDoguVersionRegistry_WatchAllCurrent_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDoguVersionRegistry_WatchAllCurrent_Call) RunAndReturn(run func(context.Context) (<-chan CurrentVersionsWatchResult, error)) *MockDoguVersionRegistry_WatchAllCurrent_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockDoguVersionRegistry creates a new instance of MockDoguVersionRegistry. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockDoguVersionRegistry(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDoguVersionRegistry {
	mock := &MockDoguVersionRegistry{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
