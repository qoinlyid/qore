package httpmw

import (
	"github.com/qoinlyid/qore"
)

// TraceID returns a X-TRACE-ID middleware.
func TraceID(next qore.HttpHandler) qore.HttpHandler {
	return func(c qore.HttpContext) error {
		xTraceID := c.Request().Header.Get(qore.HTTP_HEADER_TRACE_ID)
		if xTraceID == "" {
			xTraceID, _ = qore.StringAlphaNumRandom(32)
		}
		c.Set(qore.HTTP_CONTEXT_TRACE_ID, xTraceID)
		c.Response().Header().Set(qore.HTTP_HEADER_TRACE_ID, xTraceID)
		return next(c)
	}
}
