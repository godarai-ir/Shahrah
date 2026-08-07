// Package apperrors defines transport-independent application failures.
package apperrors

import "errors"

var (
	// ErrNotFound indicates that a requested application resource does not exist.
	ErrNotFound = errors.New("application resource not found")
	// ErrConflict indicates that an application invariant conflicts with persisted state.
	ErrConflict = errors.New("application resource conflict")
)
