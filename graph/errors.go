package graph

import "errors"

var (
	ErrNodeExists    = errors.New("graph: node already registered")
	ErrEntryRequired = errors.New("graph: entry point is required")
	ErrNodeNotFound  = errors.New("graph: node not found")
)
