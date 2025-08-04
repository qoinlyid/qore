package qore

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

// HttpContext represents the context of the current HTTP(s) request.
// It holds request and response objects, path, path parameters, data and registered handler.
type httpContextImpl struct {
	echo.Context
	server *httpServer
	logger *logger
}

// ValidateRequest.
func (h *httpContextImpl) ValidateRequest(i any) *HttpValidatorErr {
	if h.server.validator == nil {
		return &HttpValidatorErr{ErrMandatory: ErrHttpValidatorNotRegistered}
	}
	return h.server.validator.Validate(i)
}

// TraceID.
func (h *httpContextImpl) TraceID() (val string) {
	x := h.Get(HTTP_CONTEXT_TRACE_ID)
	if x != nil {
		val = fmt.Sprintf("%v", x)
	}
	return
}

// Log.
func (h *httpContextImpl) Log() *logger {
	traceID := h.TraceID()
	if traceID == "" {
		return h.logger
	}
	return h.logger.With("traceId", traceID)
}

// Api return HTTP(s) API responder.
func (h *httpContextImpl) Api() ApiResponse {
	return h.server.iApiResponse.New(h)
}

// httpHandlerToEchoHandler converts a custom HttpHandler into an Echo-compatible handler.
// It wraps echo.Context into a HttpContext to abstract away Echo internals.
func httpHandlerToEchoHandler(handler HttpHandler, server *httpServer, logger *logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handler(&httpContextImpl{
			Context: c,
			server:  server,
			logger:  logger,
		})
	}
}

// echoHandlerToHttpHandler converts an Echo handler into a custom HttpHandler.
// It unwraps HttpContext to access the underlying echo.Context.
func echoHandlerToHttpHandler(handler echo.HandlerFunc) HttpHandler {
	return func(c HttpContext) error {
		if ctx, ok := c.(*httpContextImpl); ok {
			return handler(ctx)
		}
		return echo.NewHTTPError(echo.ErrInternalServerError.Code, "invalid context implementation")
	}
}

// httpMiddlewareWrappers converts a custom HttpMiddleware(s) into an Echo-compatible middleware(s).
// It bridges custom HttpHandler-based middleware with Echo's middleware chain.
func httpMiddlewareWrappers(server *httpServer, logger *logger, middlewares ...HttpMiddleware) (echoMiddlewares []echo.MiddlewareFunc) {
	for _, m := range middlewares {
		echoMiddlewares = append(echoMiddlewares, func(next echo.HandlerFunc) echo.HandlerFunc {
			return httpHandlerToEchoHandler(
				m(echoHandlerToHttpHandler(next)),
				server,
				logger,
			)
		})
	}
	return
}
