package errors

import "errors"

type errorType int

const (
	_ErrNotFound errorType = iota + 1
	_ErrConflict
	_ErrConnection
	_ErrAlreadyExists
	_ErrGeneric
)

var _ error = Error{}

type Error struct {
	errType errorType
	cause   error
}

func (c Error) Error() string {
	return c.cause.Error()
}

func NewGenericError(err error) Error {
	return Error{
		errType: _ErrGeneric,
		cause:   err,
	}
}

func NewNotFoundError(err error) Error {
	return Error{
		errType: _ErrNotFound,
		cause:   err,
	}
}

func NewConflictError(err error) Error {
	return Error{
		errType: _ErrConflict,
		cause:   err,
	}
}

func NewConnectionError(err error) Error {
	return Error{
		errType: _ErrConnection,
		cause:   err,
	}
}

func NewAlreadyExistsError(err error) Error {
	return Error{
		errType: _ErrAlreadyExists,
		cause:   err,
	}
}

func isError(err error, t errorType) bool {
	var e Error
	if ok := errors.As(err, &e); !ok {
		return false
	}

	if e.errType == t {
		return true
	}

	return false
}

func IsGenericError(err error) bool {
	return isError(err, _ErrGeneric)
}

func IsNotFoundError(err error) bool {
	return isError(err, _ErrNotFound)
}

func IsConflictError(err error) bool {
	return isError(err, _ErrConflict)
}

func IsConnectionError(err error) bool {
	return isError(err, _ErrConnection)
}

func IsAlreadyExistsError(err error) bool {
	return isError(err, _ErrAlreadyExists)
}
