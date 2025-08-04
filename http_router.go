package qore

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// HttpRouter defines router for the HTTP server.
type HttpRouter struct {
	server *httpServer
	group  *echo.Group
	logger *logger
}

func (router *HttpRouter) path(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}

func (router *HttpRouter) clone() *HttpRouter {
	c := *router
	return &c
}

// Group creates a new router group with prefix.
func (r *HttpRouter) Group(prefix string, fn func(group *HttpRouter), middlewares ...HttpMiddleware) {
	if ValidationIsEmpty(prefix) {
		return
	}
	c := r.clone()
	c.group = r.server.core.Group(prefix, httpMiddlewareWrappers(r.server, r.logger, middlewares...)...)
	fn(c)
}

// Connect registers a new CONNECT route for a path.
func (r *HttpRouter) Connect(path string, handler HttpHandlerChain[HttpRequestPayload], middlewares ...HttpMiddleware) {
	path = r.path(path)
	if r.group != nil {
		r.group.CONNECT(
			path,
			httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
			httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
		)
		return
	}
	r.server.core.CONNECT(
		path,
		httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
		httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
	)
}

// Delete registers a new DELETE route for a path.
func (r *HttpRouter) Delete(path string, handler HttpHandlerChain[HttpRequestPayload], middlewares ...HttpMiddleware) {
	path = r.path(path)
	if r.group != nil {
		r.group.DELETE(
			path,
			httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
			httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
		)
		return
	}
	r.server.core.DELETE(
		path,
		httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
		httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
	)
}

// Get registers a new GET route for a path.
func (r *HttpRouter) Get(path string, handler HttpHandlerChain[HttpRequestPayload], middlewares ...HttpMiddleware) {
	path = r.path(path)
	if r.group != nil {
		r.group.GET(
			path,
			httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
			httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
		)
		return
	}
	r.server.core.GET(
		path,
		httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
		httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
	)
}

// Head registers a new HEAD route for a path.
func (r *HttpRouter) Head(path string, handler HttpHandlerChain[HttpRequestPayload], middlewares ...HttpMiddleware) {
	path = r.path(path)
	if r.group != nil {
		r.group.HEAD(
			path,
			httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
			httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
		)
		return
	}
	r.server.core.HEAD(
		path,
		httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
		httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
	)
}

// Options registers a new OPTIONS route for a path.
func (r *HttpRouter) Options(path string, handler HttpHandlerChain[HttpRequestPayload], middlewares ...HttpMiddleware) {
	path = r.path(path)
	if r.group != nil {
		r.group.OPTIONS(
			path,
			httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
			httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
		)
		return
	}
	r.server.core.OPTIONS(
		path,
		httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
		httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
	)
}

// Patch registers a new PATCH route for a path.
func (r *HttpRouter) Patch(path string, handler HttpHandlerChain[HttpRequestPayload], middlewares ...HttpMiddleware) {
	path = r.path(path)
	if r.group != nil {
		r.group.PATCH(
			path,
			httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
			httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
		)
		return
	}
	r.server.core.PATCH(
		path,
		httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
		httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
	)
}

// Post registers a new POST route for a path.
func (r *HttpRouter) Post(path string, handler HttpHandlerChain[HttpRequestPayload], middlewares ...HttpMiddleware) {
	path = r.path(path)
	if r.group != nil {
		r.group.POST(
			path,
			httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
			httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
		)
		return
	}
	r.server.core.POST(
		path,
		httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
		httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
	)
}

// Put registers a new PUT route for a path.
func (r *HttpRouter) Put(path string, handler HttpHandlerChain[HttpRequestPayload], middlewares ...HttpMiddleware) {
	path = r.path(path)
	if r.group != nil {
		r.group.PUT(
			path,
			httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
			httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
		)
		return
	}
	r.server.core.PUT(
		path,
		httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
		httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
	)
}

// Trace registers a new TRACE route for a path.
func (r *HttpRouter) Trace(path string, handler HttpHandlerChain[HttpRequestPayload], middlewares ...HttpMiddleware) {
	path = r.path(path)
	if r.group != nil {
		r.group.TRACE(
			path,
			httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
			httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
		)
		return
	}
	r.server.core.TRACE(
		path,
		httpHandlerToEchoHandler(handler.HandlerWrapper, r.server, r.logger),
		httpMiddlewareWrappers(r.server, r.logger, middlewares...)...,
	)
}
