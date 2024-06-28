// Code generated by mockery v2.42.1. DO NOT EDIT.

package config

import (
	io "io"

	mock "github.com/stretchr/testify/mock"
)

// MockConverter is an autogenerated mock type for the Converter type
type MockConverter struct {
	mock.Mock
}

type MockConverter_Expecter struct {
	mock *mock.Mock
}

func (_m *MockConverter) EXPECT() *MockConverter_Expecter {
	return &MockConverter_Expecter{mock: &_m.Mock}
}

// Read provides a mock function with given fields: reader
func (_m *MockConverter) Read(reader io.Reader) (Entries, error) {
	ret := _m.Called(reader)

	if len(ret) == 0 {
		panic("no return Value specified for Read")
	}

	var r0 Entries
	var r1 error
	if rf, ok := ret.Get(0).(func(io.Reader) (Entries, error)); ok {
		return rf(reader)
	}
	if rf, ok := ret.Get(0).(func(io.Reader) Entries); ok {
		r0 = rf(reader)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Entries)
		}
	}

	if rf, ok := ret.Get(1).(func(io.Reader) error); ok {
		r1 = rf(reader)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockConverter_Read_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Read'
type MockConverter_Read_Call struct {
	*mock.Call
}

// Read is a helper method to define mock.On call
//   - reader io.Reader
func (_e *MockConverter_Expecter) Read(reader interface{}) *MockConverter_Read_Call {
	return &MockConverter_Read_Call{Call: _e.mock.On("Read", reader)}
}

func (_c *MockConverter_Read_Call) Run(run func(reader io.Reader)) *MockConverter_Read_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(io.Reader))
	})
	return _c
}

func (_c *MockConverter_Read_Call) Return(_a0 Entries, _a1 error) *MockConverter_Read_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockConverter_Read_Call) RunAndReturn(run func(io.Reader) (Entries, error)) *MockConverter_Read_Call {
	_c.Call.Return(run)
	return _c
}

// Write provides a mock function with given fields: writer, cfgData
func (_m *MockConverter) Write(writer io.Writer, cfgData Entries) error {
	ret := _m.Called(writer, cfgData)

	if len(ret) == 0 {
		panic("no return Value specified for Write")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(io.Writer, Entries) error); ok {
		r0 = rf(writer, cfgData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockConverter_Write_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Write'
type MockConverter_Write_Call struct {
	*mock.Call
}

// Write is a helper method to define mock.On call
//   - writer io.Writer
//   - cfgData Entries
func (_e *MockConverter_Expecter) Write(writer interface{}, cfgData interface{}) *MockConverter_Write_Call {
	return &MockConverter_Write_Call{Call: _e.mock.On("Write", writer, cfgData)}
}

func (_c *MockConverter_Write_Call) Run(run func(writer io.Writer, cfgData Entries)) *MockConverter_Write_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(io.Writer), args[1].(Entries))
	})
	return _c
}

func (_c *MockConverter_Write_Call) Return(_a0 error) *MockConverter_Write_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConverter_Write_Call) RunAndReturn(run func(io.Writer, Entries) error) *MockConverter_Write_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockConverter creates a new instance of MockConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T Value.
func NewMockConverter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConverter {
	mock := &MockConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
