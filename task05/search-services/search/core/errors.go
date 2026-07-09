package core

import "errors"

var ErrBadArguments = errors.New("arguments are not acceptable")
var ErrNotFound = errors.New("resource is not found")
var ErrServiceUnavailable = errors.New("service unavailable")
var ErrInternal = errors.New("internal error")
