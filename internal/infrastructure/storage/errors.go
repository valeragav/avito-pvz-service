package storage

import "errors"

var (
	ErrBuildQuery    = errors.New("failed to build SQL query")
	ErrExecuteQuery  = errors.New("failed to execute query")
	ErrScanResult    = errors.New("failed to scan result")
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)
