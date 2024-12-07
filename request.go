package fetch

import (
	"context"
)

// Request can be used in ApplyFunc as a wrapper
// for the input entity to access http attributes.
type Request[T any] struct {
	Context    context.Context
	PathValues map[string]string
	Headers    map[string]string
	Body       T
}
