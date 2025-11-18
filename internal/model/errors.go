package model

import "errors"

var (
	ErrInvalidID    = errors.New("invalid courier id")
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("courier not found")
	ErrConflict     = errors.New("courier already exists")
)
