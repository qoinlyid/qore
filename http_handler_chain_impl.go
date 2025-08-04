package qore

import (
	"errors"
	"fmt"
	"net/http"
)

type httpHandlerChainImpl[T HttpRequestPayload] struct {
	requestPayload     T
	handler            HttpHandlerable
	handlerWithPayload HttpHandlerableWithPayload[T]
}

// Compile time check `httpHandlerChainImpl“ implements `HttpHandlerChain`.
var _ HttpHandlerChain[HttpRequestPayload] = (*httpHandlerChainImpl[HttpRequestPayload])(nil)

// HttpHanlderChain returns `HttpHandlerChain` that wrap user's handler.
func HttpHanlderChain(h HttpHandlerable) *httpHandlerChainImpl[HttpRequestPayload] {
	return &httpHandlerChainImpl[HttpRequestPayload]{handler: h}
}

// HttpHanlderChainWithPayload returns `HttpHandlerChain` that wrap user's handler using given payload type.
func HttpHanlderChainWithPayload[T HttpRequestPayload](h HttpHandlerableWithPayload[T]) *httpHandlerChainImpl[T] {
	var zero T
	return &httpHandlerChainImpl[T]{requestPayload: zero, handlerWithPayload: h}
}

func (chain *httpHandlerChainImpl[T]) HandlerWrapper(c HttpContext) error {
	// Check both handler.
	if chain.handler == nil && chain.handlerWithPayload == nil {
		return c.Api().ServerError(
			HttpStatusNotImplemented, errors.New("error HTTP(s) handler must be defined"),
		).Response()
	}

	// Prioritize basic handler without payload.
	if chain.handler != nil {
		return chain.handler(c)
	}

	// Then, handler with payload.
	if any(chain.requestPayload) == nil {
		return c.Api().ServerError(
			HttpStatusNotImplemented, errors.New("error Malfunction on wrap up the payload"),
		).Response()
	} else {
		// Do binding data.
		if err := c.Bind(&chain.requestPayload); err != nil {
			return c.Api().ClientError(
				HttpStatusBadRequest, errors.New("error Invalid formating on request payload"),
			).Response()
		}
		// Global validation on request payload.
		if err := c.ValidateRequest(chain.requestPayload); err != nil {
			defer err.Close()
			e := fmt.Errorf("%s, %s", http.StatusText(int(HttpStatusBadRequest)), err.Error())
			return c.Api().ClientError(HttpStatusBadRequest, e).Response()
		}
		// Custom validation from user on request payload.
		if err := chain.requestPayload.Validate(); err != nil {
			e := fmt.Errorf("%s, %s", http.StatusText(int(HttpStatusBadRequest)), err.Error())
			return c.Api().ClientError(HttpStatusBadRequest, e).Response()
		}
	}
	return chain.handlerWithPayload(c, chain.requestPayload)
}
