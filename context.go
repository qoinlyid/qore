package qore

import "context"

type contextKey string

const (
	// Context key for Trace ID.
	CTX_TRACE_ID = contextKey("traceId")
)

// ContextFromHttp wraps HTTP(s) request context and return `context.Context`.
func ContextFromHttp(c HttpContext) (ctx context.Context) {
	ctx = c.Request().Context()

	// Trace ID.
	traceId := c.TraceID()
	if !ValidationIsEmpty(traceId) {
		ctx = context.WithValue(ctx, CTX_TRACE_ID, traceId)
	}

	return
}

// ContextGetTraceID returns value of Trace ID from context.
func ContextGetTraceID(ctx context.Context) string {
	val, ok := ctx.Value(CTX_TRACE_ID).(string)
	if !ok {
		return "unknown"
	}
	return val
}
