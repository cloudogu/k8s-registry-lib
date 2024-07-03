// Code generated by mockery v2.42.1. DO NOT EDIT.

package repository

import (
	context "context"

	config "github.com/cloudogu/k8s-registry-lib/config"

	mock "github.com/stretchr/testify/mock"
)

// mockGeneralConfigRepository is an autogenerated mock type for the generalConfigRepository type
type mockGeneralConfigRepository struct {
	mock.Mock
}

type mockGeneralConfigRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *mockGeneralConfigRepository) EXPECT() *mockGeneralConfigRepository_Expecter {
	return &mockGeneralConfigRepository_Expecter{mock: &_m.Mock}
}

// create provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *mockGeneralConfigRepository) create(_a0 context.Context, _a1 configName, _a2 config.SimpleDoguName, _a3 config.Config) (config.Config, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	if len(ret) == 0 {
		panic("no return value specified for create")
	}

	var r0 config.Config
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, configName, config.SimpleDoguName, config.Config) (config.Config, error)); ok {
		return rf(_a0, _a1, _a2, _a3)
	}
	if rf, ok := ret.Get(0).(func(context.Context, configName, config.SimpleDoguName, config.Config) config.Config); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Get(0).(config.Config)
	}

	if rf, ok := ret.Get(1).(func(context.Context, configName, config.SimpleDoguName, config.Config) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockGeneralConfigRepository_create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'create'
type mockGeneralConfigRepository_create_Call struct {
	*mock.Call
}

// create is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 configName
//   - _a2 config.SimpleDoguName
//   - _a3 config.Config
func (_e *mockGeneralConfigRepository_Expecter) create(_a0 interface{}, _a1 interface{}, _a2 interface{}, _a3 interface{}) *mockGeneralConfigRepository_create_Call {
	return &mockGeneralConfigRepository_create_Call{Call: _e.mock.On("create", _a0, _a1, _a2, _a3)}
}

func (_c *mockGeneralConfigRepository_create_Call) Run(run func(_a0 context.Context, _a1 configName, _a2 config.SimpleDoguName, _a3 config.Config)) *mockGeneralConfigRepository_create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(configName), args[2].(config.SimpleDoguName), args[3].(config.Config))
	})
	return _c
}

func (_c *mockGeneralConfigRepository_create_Call) Return(_a0 config.Config, _a1 error) *mockGeneralConfigRepository_create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockGeneralConfigRepository_create_Call) RunAndReturn(run func(context.Context, configName, config.SimpleDoguName, config.Config) (config.Config, error)) *mockGeneralConfigRepository_create_Call {
	_c.Call.Return(run)
	return _c
}

// delete provides a mock function with given fields: _a0, _a1
func (_m *mockGeneralConfigRepository) delete(_a0 context.Context, _a1 configName) error {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, configName) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockGeneralConfigRepository_delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'delete'
type mockGeneralConfigRepository_delete_Call struct {
	*mock.Call
}

// delete is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 configName
func (_e *mockGeneralConfigRepository_Expecter) delete(_a0 interface{}, _a1 interface{}) *mockGeneralConfigRepository_delete_Call {
	return &mockGeneralConfigRepository_delete_Call{Call: _e.mock.On("delete", _a0, _a1)}
}

func (_c *mockGeneralConfigRepository_delete_Call) Run(run func(_a0 context.Context, _a1 configName)) *mockGeneralConfigRepository_delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(configName))
	})
	return _c
}

func (_c *mockGeneralConfigRepository_delete_Call) Return(_a0 error) *mockGeneralConfigRepository_delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockGeneralConfigRepository_delete_Call) RunAndReturn(run func(context.Context, configName) error) *mockGeneralConfigRepository_delete_Call {
	_c.Call.Return(run)
	return _c
}

// get provides a mock function with given fields: _a0, _a1
func (_m *mockGeneralConfigRepository) get(_a0 context.Context, _a1 configName) (config.Config, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for get")
	}

	var r0 config.Config
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, configName) (config.Config, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, configName) config.Config); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(config.Config)
	}

	if rf, ok := ret.Get(1).(func(context.Context, configName) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockGeneralConfigRepository_get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'get'
type mockGeneralConfigRepository_get_Call struct {
	*mock.Call
}

// get is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 configName
func (_e *mockGeneralConfigRepository_Expecter) get(_a0 interface{}, _a1 interface{}) *mockGeneralConfigRepository_get_Call {
	return &mockGeneralConfigRepository_get_Call{Call: _e.mock.On("get", _a0, _a1)}
}

func (_c *mockGeneralConfigRepository_get_Call) Run(run func(_a0 context.Context, _a1 configName)) *mockGeneralConfigRepository_get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(configName))
	})
	return _c
}

func (_c *mockGeneralConfigRepository_get_Call) Return(_a0 config.Config, _a1 error) *mockGeneralConfigRepository_get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockGeneralConfigRepository_get_Call) RunAndReturn(run func(context.Context, configName) (config.Config, error)) *mockGeneralConfigRepository_get_Call {
	_c.Call.Return(run)
	return _c
}

// saveOrMerge provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockGeneralConfigRepository) saveOrMerge(_a0 context.Context, _a1 configName, _a2 config.Config) (config.Config, error) {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for saveOrMerge")
	}

	var r0 config.Config
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, configName, config.Config) (config.Config, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, configName, config.Config) config.Config); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Get(0).(config.Config)
	}

	if rf, ok := ret.Get(1).(func(context.Context, configName, config.Config) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockGeneralConfigRepository_saveOrMerge_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'saveOrMerge'
type mockGeneralConfigRepository_saveOrMerge_Call struct {
	*mock.Call
}

// saveOrMerge is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 configName
//   - _a2 config.Config
func (_e *mockGeneralConfigRepository_Expecter) saveOrMerge(_a0 interface{}, _a1 interface{}, _a2 interface{}) *mockGeneralConfigRepository_saveOrMerge_Call {
	return &mockGeneralConfigRepository_saveOrMerge_Call{Call: _e.mock.On("saveOrMerge", _a0, _a1, _a2)}
}

func (_c *mockGeneralConfigRepository_saveOrMerge_Call) Run(run func(_a0 context.Context, _a1 configName, _a2 config.Config)) *mockGeneralConfigRepository_saveOrMerge_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(configName), args[2].(config.Config))
	})
	return _c
}

func (_c *mockGeneralConfigRepository_saveOrMerge_Call) Return(_a0 config.Config, _a1 error) *mockGeneralConfigRepository_saveOrMerge_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockGeneralConfigRepository_saveOrMerge_Call) RunAndReturn(run func(context.Context, configName, config.Config) (config.Config, error)) *mockGeneralConfigRepository_saveOrMerge_Call {
	_c.Call.Return(run)
	return _c
}

// update provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *mockGeneralConfigRepository) update(_a0 context.Context, _a1 configName, _a2 config.SimpleDoguName, _a3 config.Config) (config.Config, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	if len(ret) == 0 {
		panic("no return value specified for update")
	}

	var r0 config.Config
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, configName, config.SimpleDoguName, config.Config) (config.Config, error)); ok {
		return rf(_a0, _a1, _a2, _a3)
	}
	if rf, ok := ret.Get(0).(func(context.Context, configName, config.SimpleDoguName, config.Config) config.Config); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Get(0).(config.Config)
	}

	if rf, ok := ret.Get(1).(func(context.Context, configName, config.SimpleDoguName, config.Config) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockGeneralConfigRepository_update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'update'
type mockGeneralConfigRepository_update_Call struct {
	*mock.Call
}

// update is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 configName
//   - _a2 config.SimpleDoguName
//   - _a3 config.Config
func (_e *mockGeneralConfigRepository_Expecter) update(_a0 interface{}, _a1 interface{}, _a2 interface{}, _a3 interface{}) *mockGeneralConfigRepository_update_Call {
	return &mockGeneralConfigRepository_update_Call{Call: _e.mock.On("update", _a0, _a1, _a2, _a3)}
}

func (_c *mockGeneralConfigRepository_update_Call) Run(run func(_a0 context.Context, _a1 configName, _a2 config.SimpleDoguName, _a3 config.Config)) *mockGeneralConfigRepository_update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(configName), args[2].(config.SimpleDoguName), args[3].(config.Config))
	})
	return _c
}

func (_c *mockGeneralConfigRepository_update_Call) Return(_a0 config.Config, _a1 error) *mockGeneralConfigRepository_update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockGeneralConfigRepository_update_Call) RunAndReturn(run func(context.Context, configName, config.SimpleDoguName, config.Config) (config.Config, error)) *mockGeneralConfigRepository_update_Call {
	_c.Call.Return(run)
	return _c
}

// newMockGeneralConfigRepository creates a new instance of mockGeneralConfigRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockGeneralConfigRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockGeneralConfigRepository {
	mock := &mockGeneralConfigRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
