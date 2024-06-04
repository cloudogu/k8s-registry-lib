// Code generated by mockery v2.42.1. DO NOT EDIT.

package global

import mock "github.com/stretchr/testify/mock"

// mockEtcdConfigContext is an autogenerated mock type for the etcdConfigContext type
type mockEtcdConfigContext struct {
	mock.Mock
}

type mockEtcdConfigContext_Expecter struct {
	mock *mock.Mock
}

func (_m *mockEtcdConfigContext) EXPECT() *mockEtcdConfigContext_Expecter {
	return &mockEtcdConfigContext_Expecter{mock: &_m.Mock}
}

// Delete provides a mock function with given fields: key
func (_m *mockEtcdConfigContext) Delete(key string) error {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEtcdConfigContext_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type mockEtcdConfigContext_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - key string
func (_e *mockEtcdConfigContext_Expecter) Delete(key interface{}) *mockEtcdConfigContext_Delete_Call {
	return &mockEtcdConfigContext_Delete_Call{Call: _e.mock.On("Delete", key)}
}

func (_c *mockEtcdConfigContext_Delete_Call) Run(run func(key string)) *mockEtcdConfigContext_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockEtcdConfigContext_Delete_Call) Return(_a0 error) *mockEtcdConfigContext_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdConfigContext_Delete_Call) RunAndReturn(run func(string) error) *mockEtcdConfigContext_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteRecursive provides a mock function with given fields: key
func (_m *mockEtcdConfigContext) DeleteRecursive(key string) error {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for DeleteRecursive")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEtcdConfigContext_DeleteRecursive_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteRecursive'
type mockEtcdConfigContext_DeleteRecursive_Call struct {
	*mock.Call
}

// DeleteRecursive is a helper method to define mock.On call
//   - key string
func (_e *mockEtcdConfigContext_Expecter) DeleteRecursive(key interface{}) *mockEtcdConfigContext_DeleteRecursive_Call {
	return &mockEtcdConfigContext_DeleteRecursive_Call{Call: _e.mock.On("DeleteRecursive", key)}
}

func (_c *mockEtcdConfigContext_DeleteRecursive_Call) Run(run func(key string)) *mockEtcdConfigContext_DeleteRecursive_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockEtcdConfigContext_DeleteRecursive_Call) Return(_a0 error) *mockEtcdConfigContext_DeleteRecursive_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdConfigContext_DeleteRecursive_Call) RunAndReturn(run func(string) error) *mockEtcdConfigContext_DeleteRecursive_Call {
	_c.Call.Return(run)
	return _c
}

// Exists provides a mock function with given fields: key
func (_m *mockEtcdConfigContext) Exists(key string) (bool, error) {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for Exists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (bool, error)); ok {
		return rf(key)
	}
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockEtcdConfigContext_Exists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Exists'
type mockEtcdConfigContext_Exists_Call struct {
	*mock.Call
}

// Exists is a helper method to define mock.On call
//   - key string
func (_e *mockEtcdConfigContext_Expecter) Exists(key interface{}) *mockEtcdConfigContext_Exists_Call {
	return &mockEtcdConfigContext_Exists_Call{Call: _e.mock.On("Exists", key)}
}

func (_c *mockEtcdConfigContext_Exists_Call) Run(run func(key string)) *mockEtcdConfigContext_Exists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockEtcdConfigContext_Exists_Call) Return(_a0 bool, _a1 error) *mockEtcdConfigContext_Exists_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockEtcdConfigContext_Exists_Call) RunAndReturn(run func(string) (bool, error)) *mockEtcdConfigContext_Exists_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: key
func (_m *mockEtcdConfigContext) Get(key string) (string, error) {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(key)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockEtcdConfigContext_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type mockEtcdConfigContext_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - key string
func (_e *mockEtcdConfigContext_Expecter) Get(key interface{}) *mockEtcdConfigContext_Get_Call {
	return &mockEtcdConfigContext_Get_Call{Call: _e.mock.On("Get", key)}
}

func (_c *mockEtcdConfigContext_Get_Call) Run(run func(key string)) *mockEtcdConfigContext_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockEtcdConfigContext_Get_Call) Return(_a0 string, _a1 error) *mockEtcdConfigContext_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockEtcdConfigContext_Get_Call) RunAndReturn(run func(string) (string, error)) *mockEtcdConfigContext_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetAll provides a mock function with given fields:
func (_m *mockEtcdConfigContext) GetAll() (map[string]string, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetAll")
	}

	var r0 map[string]string
	var r1 error
	if rf, ok := ret.Get(0).(func() (map[string]string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() map[string]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockEtcdConfigContext_GetAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAll'
type mockEtcdConfigContext_GetAll_Call struct {
	*mock.Call
}

// GetAll is a helper method to define mock.On call
func (_e *mockEtcdConfigContext_Expecter) GetAll() *mockEtcdConfigContext_GetAll_Call {
	return &mockEtcdConfigContext_GetAll_Call{Call: _e.mock.On("GetAll")}
}

func (_c *mockEtcdConfigContext_GetAll_Call) Run(run func()) *mockEtcdConfigContext_GetAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockEtcdConfigContext_GetAll_Call) Return(_a0 map[string]string, _a1 error) *mockEtcdConfigContext_GetAll_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockEtcdConfigContext_GetAll_Call) RunAndReturn(run func() (map[string]string, error)) *mockEtcdConfigContext_GetAll_Call {
	_c.Call.Return(run)
	return _c
}

// GetOrFalse provides a mock function with given fields: key
func (_m *mockEtcdConfigContext) GetOrFalse(key string) (bool, string, error) {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for GetOrFalse")
	}

	var r0 bool
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(string) (bool, string, error)); ok {
		return rf(key)
	}
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(string) string); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(key)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// mockEtcdConfigContext_GetOrFalse_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOrFalse'
type mockEtcdConfigContext_GetOrFalse_Call struct {
	*mock.Call
}

// GetOrFalse is a helper method to define mock.On call
//   - key string
func (_e *mockEtcdConfigContext_Expecter) GetOrFalse(key interface{}) *mockEtcdConfigContext_GetOrFalse_Call {
	return &mockEtcdConfigContext_GetOrFalse_Call{Call: _e.mock.On("GetOrFalse", key)}
}

func (_c *mockEtcdConfigContext_GetOrFalse_Call) Run(run func(key string)) *mockEtcdConfigContext_GetOrFalse_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *mockEtcdConfigContext_GetOrFalse_Call) Return(_a0 bool, _a1 string, _a2 error) *mockEtcdConfigContext_GetOrFalse_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *mockEtcdConfigContext_GetOrFalse_Call) RunAndReturn(run func(string) (bool, string, error)) *mockEtcdConfigContext_GetOrFalse_Call {
	_c.Call.Return(run)
	return _c
}

// Refresh provides a mock function with given fields: key, timeToLiveInSeconds
func (_m *mockEtcdConfigContext) Refresh(key string, timeToLiveInSeconds int) error {
	ret := _m.Called(key, timeToLiveInSeconds)

	if len(ret) == 0 {
		panic("no return value specified for Refresh")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int) error); ok {
		r0 = rf(key, timeToLiveInSeconds)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEtcdConfigContext_Refresh_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Refresh'
type mockEtcdConfigContext_Refresh_Call struct {
	*mock.Call
}

// Refresh is a helper method to define mock.On call
//   - key string
//   - timeToLiveInSeconds int
func (_e *mockEtcdConfigContext_Expecter) Refresh(key interface{}, timeToLiveInSeconds interface{}) *mockEtcdConfigContext_Refresh_Call {
	return &mockEtcdConfigContext_Refresh_Call{Call: _e.mock.On("Refresh", key, timeToLiveInSeconds)}
}

func (_c *mockEtcdConfigContext_Refresh_Call) Run(run func(key string, timeToLiveInSeconds int)) *mockEtcdConfigContext_Refresh_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(int))
	})
	return _c
}

func (_c *mockEtcdConfigContext_Refresh_Call) Return(_a0 error) *mockEtcdConfigContext_Refresh_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdConfigContext_Refresh_Call) RunAndReturn(run func(string, int) error) *mockEtcdConfigContext_Refresh_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveAll provides a mock function with given fields:
func (_m *mockEtcdConfigContext) RemoveAll() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RemoveAll")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEtcdConfigContext_RemoveAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveAll'
type mockEtcdConfigContext_RemoveAll_Call struct {
	*mock.Call
}

// RemoveAll is a helper method to define mock.On call
func (_e *mockEtcdConfigContext_Expecter) RemoveAll() *mockEtcdConfigContext_RemoveAll_Call {
	return &mockEtcdConfigContext_RemoveAll_Call{Call: _e.mock.On("RemoveAll")}
}

func (_c *mockEtcdConfigContext_RemoveAll_Call) Run(run func()) *mockEtcdConfigContext_RemoveAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockEtcdConfigContext_RemoveAll_Call) Return(_a0 error) *mockEtcdConfigContext_RemoveAll_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdConfigContext_RemoveAll_Call) RunAndReturn(run func() error) *mockEtcdConfigContext_RemoveAll_Call {
	_c.Call.Return(run)
	return _c
}

// Set provides a mock function with given fields: key, value
func (_m *mockEtcdConfigContext) Set(key string, value string) error {
	ret := _m.Called(key, value)

	if len(ret) == 0 {
		panic("no return value specified for Set")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEtcdConfigContext_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type mockEtcdConfigContext_Set_Call struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//   - key string
//   - value string
func (_e *mockEtcdConfigContext_Expecter) Set(key interface{}, value interface{}) *mockEtcdConfigContext_Set_Call {
	return &mockEtcdConfigContext_Set_Call{Call: _e.mock.On("Set", key, value)}
}

func (_c *mockEtcdConfigContext_Set_Call) Run(run func(key string, value string)) *mockEtcdConfigContext_Set_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *mockEtcdConfigContext_Set_Call) Return(_a0 error) *mockEtcdConfigContext_Set_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdConfigContext_Set_Call) RunAndReturn(run func(string, string) error) *mockEtcdConfigContext_Set_Call {
	_c.Call.Return(run)
	return _c
}

// SetWithLifetime provides a mock function with given fields: key, value, timeToLiveInSeconds
func (_m *mockEtcdConfigContext) SetWithLifetime(key string, value string, timeToLiveInSeconds int) error {
	ret := _m.Called(key, value, timeToLiveInSeconds)

	if len(ret) == 0 {
		panic("no return value specified for SetWithLifetime")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, int) error); ok {
		r0 = rf(key, value, timeToLiveInSeconds)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockEtcdConfigContext_SetWithLifetime_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetWithLifetime'
type mockEtcdConfigContext_SetWithLifetime_Call struct {
	*mock.Call
}

// SetWithLifetime is a helper method to define mock.On call
//   - key string
//   - value string
//   - timeToLiveInSeconds int
func (_e *mockEtcdConfigContext_Expecter) SetWithLifetime(key interface{}, value interface{}, timeToLiveInSeconds interface{}) *mockEtcdConfigContext_SetWithLifetime_Call {
	return &mockEtcdConfigContext_SetWithLifetime_Call{Call: _e.mock.On("SetWithLifetime", key, value, timeToLiveInSeconds)}
}

func (_c *mockEtcdConfigContext_SetWithLifetime_Call) Run(run func(key string, value string, timeToLiveInSeconds int)) *mockEtcdConfigContext_SetWithLifetime_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(int))
	})
	return _c
}

func (_c *mockEtcdConfigContext_SetWithLifetime_Call) Return(_a0 error) *mockEtcdConfigContext_SetWithLifetime_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockEtcdConfigContext_SetWithLifetime_Call) RunAndReturn(run func(string, string, int) error) *mockEtcdConfigContext_SetWithLifetime_Call {
	_c.Call.Return(run)
	return _c
}

// newMockEtcdConfigContext creates a new instance of mockEtcdConfigContext. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockEtcdConfigContext(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockEtcdConfigContext {
	mock := &mockEtcdConfigContext{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}