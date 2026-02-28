package myerrors

import (
	"errors"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrHashPassword   = errors.New("error hashing password")
)
