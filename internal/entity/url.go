package entity

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrConflictURL = errors.New("short key already exists")
)
