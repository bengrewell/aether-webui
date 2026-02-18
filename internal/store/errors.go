package store

import "errors"

var (
	ErrConflict        = errors.New("store: conflict")
	ErrNotFound        = errors.New("store: not found")
	ErrExpired         = errors.New("store: expired")
	ErrInvalidArgument = errors.New("store: invalid argument")
)
