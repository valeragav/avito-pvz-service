package infra

import "errors"

var (
	ErrBuildQuery   = errors.New("failed to build SQL query")
	ErrExecuteQuery = errors.New("failed to execute query")
	ErrScanResult   = errors.New("failed to scan result")
)

var (
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("duplicate")
)
