package lsof

import (
	"errors"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrNotImplemented = errors.New("not implemented")
)
