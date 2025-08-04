package httpmw

import "github.com/qoinlyid/qore"

// MiddlewareName defines custom string type wrapper to identified middleware name.
type MiddlewareName string

// Skipper defines a function to skip middleware.
// Returning true skips processing the middleware.
type Skipper func(c qore.HttpContext) bool

// ValuesExtractor defines a function for extracting values (keys/tokens) from the given context.
type ValuesExtractor func(c qore.HttpContext) ([]string, error)

// DefaultSkipper returns false which process the middleware.
func DefaultSkipper(c qore.HttpContext) bool { return false }
