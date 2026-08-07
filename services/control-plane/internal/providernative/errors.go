// Package providernative owns provider-native resource application contracts.
package providernative

import "errors"

var (
	// ErrResourceNotFound indicates that a provider-native resource mapping does not exist.
	ErrResourceNotFound = errors.New("provider native resource not found")
)
