package qore

import "net/http"

// HttpResponseSuccess defines type of HTTP(s) status response success.
type HttpResponseSuccess int

// HTTP(s) success's status codes.
const (
	HttpStatusOK                   HttpResponseSuccess = http.StatusOK
	HttpStatusCreated              HttpResponseSuccess = http.StatusCreated
	HttpStatusAccepted             HttpResponseSuccess = http.StatusAccepted
	HttpStatusNonAuthoritativeInfo HttpResponseSuccess = http.StatusNonAuthoritativeInfo
	HttpStatusNoContent            HttpResponseSuccess = http.StatusNoContent
	HttpStatusResetContent         HttpResponseSuccess = http.StatusResetContent
	HttpStatusPartialContent       HttpResponseSuccess = http.StatusPartialContent
)

// HttpResponseRedirect defines type of HTTP(s) status response redirect.
type HttpResponseRedirect int

// HTTP(s) redirect's status codes.
const (
	HttpStatusMovedPermanently HttpResponseRedirect = http.StatusMovedPermanently
	HttpStatusFound            HttpResponseRedirect = http.StatusFound
	HttpStatusNotModified      HttpResponseRedirect = http.StatusNotModified
)

// HttpResponseClientError defines type of HTTP(s) status response client error.
type HttpResponseClientError int

// HTTP(s) client error's status codes.
const (
	HttpStatusBadRequest        HttpResponseClientError = http.StatusBadRequest
	HttpStatusUnauthorized      HttpResponseClientError = http.StatusUnauthorized
	HttpStatusPaymentRequired   HttpResponseClientError = http.StatusPaymentRequired
	HttpStatusForbidden         HttpResponseClientError = http.StatusForbidden
	HttpStatusNotFound          HttpResponseClientError = http.StatusNotFound
	HttpStatusMethodNotAllowed  HttpResponseClientError = http.StatusMethodNotAllowed
	HttpStatusNotAcceptable     HttpResponseClientError = http.StatusNotAcceptable
	HttpStatusProxyAuthRequired HttpResponseClientError = http.StatusProxyAuthRequired
	HttpStatusRequestTimeout    HttpResponseClientError = http.StatusRequestTimeout
	HttpStatusConflict          HttpResponseClientError = http.StatusConflict
	HttpStatusGone              HttpResponseClientError = http.StatusGone
)

// HttpResponseServerError defines type of HTTP(s) status response server error.
type HttpResponseServerError int

// HTTP(s) server error's status codes.
const (
	HttpStatusInternalServerError HttpResponseServerError = http.StatusInternalServerError
	HttpStatusNotImplemented      HttpResponseServerError = http.StatusNotImplemented
	HttpStatusBadGateway          HttpResponseServerError = http.StatusBadGateway
	HttpStatusServiceUnavailable  HttpResponseServerError = http.StatusServiceUnavailable
	HttpStatusGatewayTimeout      HttpResponseServerError = http.StatusGatewayTimeout
)

type Responder interface {
	Encode()
}

func (c *httpContextImpl) ResponseSuccess() error {
	// http.StatusAccepted
	return nil
}

func (c *httpContextImpl) ResponseClientError() error {
	return nil
}

func (c *httpContextImpl) ResponseServerError() error {
	return nil
}

type HttpRequestPayload interface {
	Validate() error
}
