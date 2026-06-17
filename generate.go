// Package tools centralizes code generation directives for the project.
//
// Run:  go generate ./...   (or: make mocks)
//
// Mockery reads its configuration from .mockery.yaml and regenerates all mocks
// declared there (low-level / infrastructure interfaces) into a dedicated
// `mock` sub-package next to each interface.
package tools

//go:generate go tool mockery
