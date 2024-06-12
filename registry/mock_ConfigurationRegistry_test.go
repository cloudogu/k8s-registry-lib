// Code generated by mockery v2.42.1. DO NOT EDIT.

package registry

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockConfigurationRegistry is an autogenerated mock type for the ConfigurationRegistry type
type MockConfigurationRegistry struct {
	mock.Mock
}

type MockConfigurationRegistry_Expecter struct {
	mock *mock.Mock
}

func (_m *MockConfigurationRegistry) EXPECT() *MockConfigurationRegistry_Expecter {
	return &MockConfigurationRegistry_Expecter{mock: &_m.Mock}
}

// Delete provides a mock function with given fields: ctx, key
func (_m *MockConfigurationRegistry) Delete(ctx context.Context, key string) error {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockConfigurationRegistry_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type MockConfigurationRegistry_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *MockConfigurationRegistry_Expecter) Delete(ctx interface{}, key interface{}) *MockConfigurationRegistry_Delete_Call {
	return &MockConfigurationRegistry_Delete_Call{Call: _e.mock.On("Delete", ctx, key)}
}

func (_c *MockConfigurationRegistry_Delete_Call) Run(run func(ctx context.Context, key string)) *MockConfigurationRegistry_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockConfigurationRegistry_Delete_Call) Return(_a0 error) *MockConfigurationRegistry_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConfigurationRegistry_Delete_Call) RunAndReturn(run func(context.Context, string) error) *MockConfigurationRegistry_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteAll provides a mock function with given fields: ctx
func (_m *MockConfigurationRegistry) DeleteAll(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for DeleteAll")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockConfigurationRegistry_DeleteAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteAll'
type MockConfigurationRegistry_DeleteAll_Call struct {
	*mock.Call
}

// DeleteAll is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockConfigurationRegistry_Expecter) DeleteAll(ctx interface{}) *MockConfigurationRegistry_DeleteAll_Call {
	return &MockConfigurationRegistry_DeleteAll_Call{Call: _e.mock.On("DeleteAll", ctx)}
}

func (_c *MockConfigurationRegistry_DeleteAll_Call) Run(run func(ctx context.Context)) *MockConfigurationRegistry_DeleteAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockConfigurationRegistry_DeleteAll_Call) Return(_a0 error) *MockConfigurationRegistry_DeleteAll_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConfigurationRegistry_DeleteAll_Call) RunAndReturn(run func(context.Context) error) *MockConfigurationRegistry_DeleteAll_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteRecursive provides a mock function with given fields: ctx, key
func (_m *MockConfigurationRegistry) DeleteRecursive(ctx context.Context, key string) error {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for DeleteRecursive")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockConfigurationRegistry_DeleteRecursive_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteRecursive'
type MockConfigurationRegistry_DeleteRecursive_Call struct {
	*mock.Call
}

// DeleteRecursive is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *MockConfigurationRegistry_Expecter) DeleteRecursive(ctx interface{}, key interface{}) *MockConfigurationRegistry_DeleteRecursive_Call {
	return &MockConfigurationRegistry_DeleteRecursive_Call{Call: _e.mock.On("DeleteRecursive", ctx, key)}
}

func (_c *MockConfigurationRegistry_DeleteRecursive_Call) Run(run func(ctx context.Context, key string)) *MockConfigurationRegistry_DeleteRecursive_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockConfigurationRegistry_DeleteRecursive_Call) Return(_a0 error) *MockConfigurationRegistry_DeleteRecursive_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConfigurationRegistry_DeleteRecursive_Call) RunAndReturn(run func(context.Context, string) error) *MockConfigurationRegistry_DeleteRecursive_Call {
	_c.Call.Return(run)
	return _c
}

// Exists provides a mock function with given fields: ctx, key
func (_m *MockConfigurationRegistry) Exists(ctx context.Context, key string) (bool, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Exists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (bool, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConfigurationRegistry_Exists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Exists'
type MockConfigurationRegistry_Exists_Call struct {
	*mock.Call
}

// Exists is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *MockConfigurationRegistry_Expecter) Exists(ctx interface{}, key interface{}) *MockConfigurationRegistry_Exists_Call {
	return &MockConfigurationRegistry_Exists_Call{Call: _e.mock.On("Exists", ctx, key)}
}

func (_c *MockConfigurationRegistry_Exists_Call) Run(run func(ctx context.Context, key string)) *MockConfigurationRegistry_Exists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockConfigurationRegistry_Exists_Call) Return(_a0 bool, _a1 error) *MockConfigurationRegistry_Exists_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConfigurationRegistry_Exists_Call) RunAndReturn(run func(context.Context, string) (bool, error)) *MockConfigurationRegistry_Exists_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, key
func (_m *MockConfigurationRegistry) Get(ctx context.Context, key string) (string, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConfigurationRegistry_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MockConfigurationRegistry_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *MockConfigurationRegistry_Expecter) Get(ctx interface{}, key interface{}) *MockConfigurationRegistry_Get_Call {
	return &MockConfigurationRegistry_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *MockConfigurationRegistry_Get_Call) Run(run func(ctx context.Context, key string)) *MockConfigurationRegistry_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockConfigurationRegistry_Get_Call) Return(_a0 string, _a1 error) *MockConfigurationRegistry_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConfigurationRegistry_Get_Call) RunAndReturn(run func(context.Context, string) (string, error)) *MockConfigurationRegistry_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetAll provides a mock function with given fields: ctx
func (_m *MockConfigurationRegistry) GetAll(ctx context.Context) (map[string]string, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetAll")
	}

	var r0 map[string]string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (map[string]string, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) map[string]string); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConfigurationRegistry_GetAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAll'
type MockConfigurationRegistry_GetAll_Call struct {
	*mock.Call
}

// GetAll is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockConfigurationRegistry_Expecter) GetAll(ctx interface{}) *MockConfigurationRegistry_GetAll_Call {
	return &MockConfigurationRegistry_GetAll_Call{Call: _e.mock.On("GetAll", ctx)}
}

func (_c *MockConfigurationRegistry_GetAll_Call) Run(run func(ctx context.Context)) *MockConfigurationRegistry_GetAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockConfigurationRegistry_GetAll_Call) Return(_a0 map[string]string, _a1 error) *MockConfigurationRegistry_GetAll_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConfigurationRegistry_GetAll_Call) RunAndReturn(run func(context.Context) (map[string]string, error)) *MockConfigurationRegistry_GetAll_Call {
	_c.Call.Return(run)
	return _c
}

// Set provides a mock function with given fields: ctx, key, value
func (_m *MockConfigurationRegistry) Set(ctx context.Context, key string, value string) error {
	ret := _m.Called(ctx, key, value)

	if len(ret) == 0 {
		panic("no return value specified for Set")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockConfigurationRegistry_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type MockConfigurationRegistry_Set_Call struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - value string
func (_e *MockConfigurationRegistry_Expecter) Set(ctx interface{}, key interface{}, value interface{}) *MockConfigurationRegistry_Set_Call {
	return &MockConfigurationRegistry_Set_Call{Call: _e.mock.On("Set", ctx, key, value)}
}

func (_c *MockConfigurationRegistry_Set_Call) Run(run func(ctx context.Context, key string, value string)) *MockConfigurationRegistry_Set_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockConfigurationRegistry_Set_Call) Return(_a0 error) *MockConfigurationRegistry_Set_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConfigurationRegistry_Set_Call) RunAndReturn(run func(context.Context, string, string) error) *MockConfigurationRegistry_Set_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockConfigurationRegistry creates a new instance of MockConfigurationRegistry. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockConfigurationRegistry(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConfigurationRegistry {
	mock := &MockConfigurationRegistry{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
