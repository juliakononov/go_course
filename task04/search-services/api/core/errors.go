package core

import "errors"

var ErrBadArguments = errors.New("arguments are not acceptable")
var ErrAlreadyExists = errors.New("resource or task already exists")
var ErrServiceUnavailable = errors.New("service unavailable")
var ErrInternal = errors.New("internal error")
