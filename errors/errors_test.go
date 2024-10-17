package errors

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError(assert.AnError)

	assert.Equal(t, assert.AnError, err.cause)
	assert.Equal(t, _ErrNotFound, err.errType)
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError(assert.AnError)

	assert.Equal(t, assert.AnError, err.cause)
	assert.Equal(t, _ErrConflict, err.errType)
}

func TestNewConnectionError(t *testing.T) {
	err := NewConnectionError(assert.AnError)

	assert.Equal(t, assert.AnError, err.cause)
	assert.Equal(t, _ErrConnection, err.errType)
}

func TestNewGenericError(t *testing.T) {
	err := NewGenericError(assert.AnError)

	assert.Equal(t, assert.AnError, err.cause)
	assert.Equal(t, _ErrGeneric, err.errType)
}

func TestNewWatchError(t *testing.T) {
	err := NewWatchError(assert.AnError)

	assert.Equal(t, assert.AnError, err.cause)
	assert.Equal(t, _ErrWatch, err.errType)
}

func TestError_Error(t *testing.T) {
	err := Error{
		errType: 0,
		cause:   assert.AnError,
	}

	assert.Equal(t, assert.AnError.Error(), err.Error())
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		xResult bool
	}{
		{
			name:    "NotFoundError",
			err:     NewNotFoundError(assert.AnError),
			xResult: true,
		},
		{
			name:    "ConflictError",
			err:     NewConflictError(assert.AnError),
			xResult: false,
		},
		{
			name:    "No config error",
			err:     assert.AnError,
			xResult: false,
		},
		{
			name:    "error is nil",
			err:     nil,
			xResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.xResult, IsNotFoundError(tc.err))
		})
	}
}

func TestIsConflictError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		xResult bool
	}{
		{
			name:    "ConflictError",
			err:     NewConflictError(assert.AnError),
			xResult: true,
		},
		{
			name:    "NotFoundError",
			err:     NewNotFoundError(assert.AnError),
			xResult: false,
		},
		{
			name:    "No config error",
			err:     assert.AnError,
			xResult: false,
		},
		{
			name:    "error is nil",
			err:     nil,
			xResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.xResult, IsConflictError(tc.err))
		})
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		xResult bool
	}{
		{
			name:    "ConnectionError",
			err:     NewConnectionError(assert.AnError),
			xResult: true,
		},
		{
			name:    "NotFoundError",
			err:     NewNotFoundError(assert.AnError),
			xResult: false,
		},
		{
			name:    "No config error",
			err:     assert.AnError,
			xResult: false,
		},
		{
			name:    "error is nil",
			err:     nil,
			xResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.xResult, IsConnectionError(tc.err))
		})
	}
}

func TestIsAlreadyExistsError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		xResult bool
	}{
		{
			name:    "AlreadyExistsError",
			err:     NewAlreadyExistsError(assert.AnError),
			xResult: true,
		},
		{
			name:    "NotFoundError",
			err:     NewNotFoundError(assert.AnError),
			xResult: false,
		},
		{
			name:    "No config error",
			err:     assert.AnError,
			xResult: false,
		},
		{
			name:    "error is nil",
			err:     nil,
			xResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.xResult, IsAlreadyExistsError(tc.err))
		})
	}
}

func TestIsGenericError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		xResult bool
	}{
		{
			name:    "GenericError",
			err:     NewGenericError(assert.AnError),
			xResult: true,
		},
		{
			name:    "NotFoundError",
			err:     NewNotFoundError(assert.AnError),
			xResult: false,
		},
		{
			name:    "No config error",
			err:     assert.AnError,
			xResult: false,
		},
		{
			name:    "error is nil",
			err:     nil,
			xResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.xResult, IsGenericError(tc.err))
		})
	}
}

func TestIsWatchError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		xResult bool
	}{
		{
			name:    "WatchError",
			err:     NewWatchError(assert.AnError),
			xResult: true,
		},
		{
			name:    "NotFoundError",
			err:     NewNotFoundError(assert.AnError),
			xResult: false,
		},
		{
			name:    "No config error",
			err:     assert.AnError,
			xResult: false,
		},
		{
			name:    "error is nil",
			err:     nil,
			xResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.xResult, IsWatchError(tc.err))
		})
	}
}
