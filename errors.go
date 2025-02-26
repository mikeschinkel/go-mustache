package mustache

import (
	"errors"
)

// Package-level errors that can be returned by various functions.
var (
	// ErrDynamicRefNotFound is returned when a dynamic reference (as specified in
	// the optional "dynamic-names" part of the Mustache spec) cannot be resolved
	// to a valid template name. This occurs when using the {{>data.reference}} syntax
	// and the dynamic lookup fails.
	ErrDynamicRefNotFound = errors.New("dynamic reference not found")
)
